#!/usr/bin/env python3
"""Fetch point children map with shared references.

Full accuracy at all levels. Every chain is recorded, but identical
children sets are stored once and referenced by ID.

Output:
{
  "sets": { "s1": ["base64", "gzip", ...], "s2": [...], ... },
  "chains": { "action_ext": "s1", "action_ext > base64": "s1", ... }
}

Usage: API_TOKEN=xxx python3 scripts/fetch_point_refs.py > spec/point_map.json
"""

import json
import os
import ssl
import sys
import urllib.request
from collections import deque

API_HOST = os.environ.get("API_HOST", "https://api.wallarm.com")
API_TOKEN = os.environ.get("API_TOKEN", os.environ.get("WALLARM_API_TOKEN", ""))
MAX_DEPTH = 5

INT_PAIRED = {
    "array", "grpc", "json_array", "path",
    "xml_pi", "xml_dtd_entity", "xml_tag_array", "xml_comment",
    "viewstate_array", "viewstate_pair", "viewstate_triplet",
}

STR_PAIRED = {
    "get", "header", "form_urlencoded", "cookie", "hash", "multipart",
    "jwt", "json", "json_obj", "protobuf", "content_disp", "response_header",
    "xml_tag", "xml_attr",
    "gql_query", "gql_mutation", "gql_subscription", "gql_fragment",
    "gql_dir", "gql_spread", "gql_type", "gql_var",
    "viewstate_dict", "viewstate_sparse_array",
}

_call_count = 0


def fetch_children(points):
    global _call_count
    _call_count += 1
    payload = json.dumps({"mode": "default", "points": [points]}).encode()
    req = urllib.request.Request(
        f"{API_HOST}/v2/suggest/point",
        data=payload,
        headers={
            "Content-Type": "application/json",
            "X-WallarmAPI-Token": API_TOKEN,
            "Accept": "application/json",
        },
    )
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    try:
        with urllib.request.urlopen(req, context=ctx) as resp:
            data = json.loads(resp.read())
            body = data.get("body", [[]])
            return sorted(body[0]) if body and body[0] else []
    except Exception as e:
        print(f"ERROR [{_call_count}] {e}", file=sys.stderr)
        return []


def make_point(element):
    if element in INT_PAIRED:
        return [element, 0]
    elif element in STR_PAIRED:
        return [element, "any"]
    return [element]


def chain_key(points):
    parts = []
    for p in points:
        if len(p) > 1:
            parts.append(f"{p[0]}:{p[1]}")
        else:
            parts.append(p[0])
    return " > ".join(parts)


def main():
    if not API_TOKEN:
        print("Set API_TOKEN or WALLARM_API_TOKEN env var", file=sys.stderr)
        sys.exit(1)

    # Children sets: tuple → set_id
    sets_by_tuple = {}
    sets_by_id = {}
    set_counter = 0

    # Chain map: chain_key → set_id (or "leaf")
    chains = {}

    # Track which children sets we've already recursed into
    explored_sets = set()

    queue = deque()

    base_points = [
        "action_ext", "action_name",
        "get", "get_all", "get_name",
        "header", "header_all", "header_name",
        "path", "path_all",
        "post", "uri",
    ]

    for bp in base_points:
        queue.append(([make_point(bp)], 1))

    print(f"Fetching with references (max depth {MAX_DEPTH})...", file=sys.stderr, flush=True)

    while queue:
        chain, depth = queue.popleft()
        key = chain_key(chain)

        if key in chains:
            continue

        children = fetch_children(chain)

        if not children:
            chains[key] = "leaf"
            print(f"  [{_call_count}] d={depth} {key} → leaf", file=sys.stderr, flush=True)
            continue

        # Get or create set ID for this children list
        children_tuple = tuple(children)
        if children_tuple not in sets_by_tuple:
            set_counter += 1
            set_id = f"s{set_counter}"
            sets_by_tuple[children_tuple] = set_id
            sets_by_id[set_id] = children
        else:
            set_id = sets_by_tuple[children_tuple]

        chains[key] = set_id
        print(f"  [{_call_count}] d={depth} {key} → {set_id} ({len(children)} children)", file=sys.stderr, flush=True)

        if depth >= MAX_DEPTH:
            continue

        # Only recurse into children if this SET hasn't been explored yet
        # (all chains referencing the same set would produce the same subtree)
        if children_tuple in explored_sets:
            continue
        explored_sets.add(children_tuple)

        for child in children:
            next_chain = chain + [make_point(child)]
            next_key = chain_key(next_chain)
            if next_key not in chains:
                queue.append((next_chain, depth + 1))

    print(f"\nTotal API calls: {_call_count}", file=sys.stderr, flush=True)
    print(f"Total chains: {len(chains)}", file=sys.stderr, flush=True)
    print(f"  with children: {sum(1 for v in chains.values() if v != 'leaf')}", file=sys.stderr, flush=True)
    print(f"  leaves: {sum(1 for v in chains.values() if v == 'leaf')}", file=sys.stderr, flush=True)
    print(f"Unique children sets: {len(sets_by_id)}", file=sys.stderr, flush=True)

    output = {
        "sets": sets_by_id,
        "chains": chains,
    }

    json.dump(output, sys.stdout, indent=2, sort_keys=True)
    print()


if __name__ == "__main__":
    main()
