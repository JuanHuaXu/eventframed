"""Freeze paraphrased official HTTP facts separately from retrieval outputs."""

import json
from pathlib import Path


ROOT = Path(__file__).parent
HTTP = "https://www.rfc-editor.org/rfc/rfc9110.html"
CACHE = "https://www.rfc-editor.org/rfc/rfc9111.html"

# id, event-frame text, literal query, paraphrase query, official section URL
FACTS = [
    ("get", "HTTP GET requests a current selected representation of the target resource.",
     "Which HTTP method requests the current representation of a target resource?",
     "Which operation asks the server to send what a resource currently contains?", HTTP + "#section-9.3.1"),
    ("head", "HTTP HEAD has GET-like semantics but the response omits content and can provide representation metadata.",
     "Which HTTP method has GET-like semantics without response content?",
     "Which operation checks a resource's headers without downloading its body?", HTTP + "#section-9.3.2"),
    ("post", "HTTP POST asks a target resource to process the enclosed representation according to that resource's own semantics.",
     "Which HTTP method asks a target to process submitted representation data?",
     "Which operation lets the receiving resource decide how to handle submitted data?", HTTP + "#section-9.3.3"),
    ("put", "HTTP PUT requests creation or replacement of the target resource state using the enclosed representation.",
     "Which HTTP method creates or replaces target resource state with the enclosed representation?",
     "Which operation installs supplied content as the target's new state?", HTTP + "#section-9.3.4"),
    ("delete", "HTTP DELETE requests removal of the association between a target resource and its current functionality.",
     "Which HTTP method requests removal of a target resource's current association?",
     "Which operation asks the server to unlink a resource from its present behavior?", HTTP + "#section-9.3.5"),
    ("connect", "HTTP CONNECT requests a tunnel to the destination origin server identified by the target.",
     "Which HTTP method requests a tunnel to the destination origin server?",
     "Which operation turns the connection into a conduit toward the target server?", HTTP + "#section-9.3.6"),
    ("options", "HTTP OPTIONS requests information about communication options available for a target resource.",
     "Which HTTP method asks about available communication choices for the target resource?",
     "Which operation asks a resource what ways of communicating it supports?", HTTP + "#section-9.3.7"),
    ("trace", "HTTP TRACE requests a remote loop-back of the request message along the path to the target.",
     "Which HTTP method requests remote loop-back of the request message?",
     "Which operation asks the far end to echo the received request?", HTTP + "#section-9.3.8"),
    ("405", "HTTP 405 Method Not Allowed means the origin knows the request method but the target resource does not support it.",
     "Which HTTP status means a known request method is not allowed for this target resource?",
     "Which response code says the server recognizes the operation but this resource rejects it?", HTTP + "#section-15.5.6"),
    ("501", "HTTP 501 Not Implemented means the server lacks functionality required to fulfill the request.",
     "Which HTTP status means the server lacks required functionality to fulfill a request?",
     "Which response code says the server does not implement the needed operation?", HTTP + "#section-15.6.2"),
    ("204", "HTTP 204 No Content means the request succeeded and there is no additional response content to send.",
     "Which HTTP status reports success with no additional response content?",
     "Which success code says the action worked but the reply has no body?", HTTP + "#section-15.3.5"),
    ("304", "HTTP 304 Not Modified tells a conditional GET or HEAD client that a stored representation can be reused.",
     "Which HTTP status says a conditional GET or HEAD found no modification?",
     "Which response code lets a conditional fetch reuse its stored copy?", HTTP + "#section-15.4.5"),
    ("no-store", "The HTTP no-store response directive normally forbids a cache from storing any part of the immediate request or response.",
     "Which Cache-Control response directive tells a cache not to store the request or response?",
     "Which caching instruction says not to retain either side of this exchange for later use?", CACHE + "#section-5.2.2.5"),
    ("private", "The unqualified HTTP private response directive forbids a shared cache from storing the response while permitting a private cache subject to other rules.",
     "Which Cache-Control response directive forbids shared-cache storage but may permit user-local storage?",
     "Which caching instruction keeps an unqualified reply in a user-local cache rather than a shared one?", CACHE + "#section-5.2.2.7"),
    ("must-revalidate", "The HTTP must-revalidate response directive forbids reuse of a stale response until successful origin validation.",
     "Which Cache-Control response directive requires origin validation before reusing a stale response?",
     "Which caching instruction demands a fresh check with the original server before serving old data?", CACHE + "#section-5.2.2.2"),
]

corpus = json.loads((ROOT / "prepared-v2/corpus.json").read_text())
queries = []
oracle = {}
for key, text, literal, paraphrase, source in FACTS:
    fixture = f"rfc-{key}"
    corpus.append({"fixture_id": fixture, "text": text, "source": source})
    for wording, question in (("literal", literal), ("paraphrase", paraphrase)):
        case = f"{fixture}-{wording}"
        queries.append({"case_id": case, "split": "confirmation", "question": question})
        oracle[case] = {"answer": key, "support": [fixture], "cluster": fixture, "wording": wording}
for key, question in (
    ("patch", "Which retained HTTP method named PATCH describes partial modifications?"),
    ("immutable", "Which retained Cache-Control directive named immutable says a fresh response will not change?"),
):
    case = f"rfc-absent-{key}"
    queries.append({"case_id": case, "split": "confirmation", "question": question})
    oracle[case] = {"answer": "UNKNOWN", "support": [], "cluster": case, "wording": "absent"}

out = ROOT / "rfc-transfer-v1"
out.mkdir(exist_ok=False)
for name, value in (("corpus", corpus), ("queries", queries), ("oracle", oracle)):
    with (out / f"{name}.json").open("x") as handle:
        json.dump(value, handle, indent=2)
        handle.write("\n")
print(f"{len(corpus)} corpus records, {len(queries)} queries, {len(FACTS)} new fact clusters")
