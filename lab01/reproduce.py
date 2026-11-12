#!/usr/bin/env python3
"""Exercise the local OAuth flow and detect calendar field over-access."""

import argparse
from html.parser import HTMLParser
import json
import secrets
import ssl
import string
from http.cookiejar import CookieJar
from urllib.error import HTTPError
from urllib.parse import urlencode
from urllib.request import (
    HTTPCookieProcessor,
    HTTPSHandler,
    ProxyHandler,
    Request,
    build_opener,
)


class ConsentParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.hidden = {}

    def handle_starttag(self, tag, attrs):
        if tag != "input":
            return
        attributes = dict(attrs)
        if attributes.get("type") != "hidden":
            return
        name = attributes.get("name")
        if name:
            self.hidden[name] = attributes.get("value", "")


def random_username():
    alphabet = string.ascii_lowercase + string.digits
    return "lab" + "".join(secrets.choice(alphabet) for _ in range(12))


def run(variant):
    server = "https://server-01.oauth.labs"
    client = "https://client-01.oauth.labs"
    context = ssl.create_default_context()
    context.check_hostname = False
    context.verify_mode = ssl.CERT_NONE
    opener = build_opener(
        ProxyHandler({}),
        HTTPCookieProcessor(CookieJar()),
        HTTPSHandler(context=context),
    )

    def fetch(url, form=None):
        data = urlencode(form).encode() if form is not None else None
        request = Request(
            url,
            data=data,
            headers={"User-Agent": f"lab01-reproduce/{variant}"},
        )
        with opener.open(request, timeout=15) as response:
            return response.status, response.geturl(), response.read().decode("utf-8")

    username = random_username()
    password = secrets.token_urlsafe(18)
    fetch(server + "/register", {"username": username, "password": password})
    fetch(server + "/login", {"username": username, "password": password})

    _, _, login_page = fetch(client + "/login?task=calendar.freebusy")
    authorization_url = None
    for match in _anchor_hrefs(login_page):
        if "/oauth/authorize?" in match:
            authorization_url = match
            break
    if not authorization_url:
        raise RuntimeError("client login page did not contain an OAuth authorization URL")

    _, _, consent_page = fetch(authorization_url)
    consent = ConsentParser()
    consent.feed(consent_page)
    if consent.hidden.get("task_id") != "calendar.freebusy":
        raise RuntimeError("task consent page did not bind the requested task")

    try:
        _, _, final_page = fetch(server + "/consents", consent.hidden)
    except HTTPError as error:
        body = error.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"consent request failed with {error.code}: {body}") from error

    title_visible = "Title:</strong> Security project review" in final_page
    description_visible = "Description:</strong> Review the threat model" in final_page
    exploitable = title_visible and description_visible
    result = {
        "exploitable": exploitable,
        "impact_verified": exploitable,
    }
    print(json.dumps(result))


def _anchor_hrefs(html):
    class AnchorParser(HTMLParser):
        def __init__(self):
            super().__init__()
            self.hrefs = []

        def handle_starttag(self, tag, attrs):
            if tag == "a":
                href = dict(attrs).get("href")
                if href:
                    self.hrefs.append(href)

    parser = AnchorParser()
    parser.feed(html)
    return parser.hrefs


if __name__ == "__main__":
    argument_parser = argparse.ArgumentParser(description=__doc__)
    argument_parser.add_argument("--variant", required=True, choices=("vulnerable", "fixed"))
    arguments = argument_parser.parse_args()
    run(arguments.variant)
