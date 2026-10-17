#!/usr/bin/env python3
"""Reset the GemBid database to a clean demo state.

DESTRUCTIVE: empties EVERY table (schema is kept) and restarts IDs at 1,
then creates demo accounts and data through the real API.

Run on the server, from the repo root, with the backend running:

    set -a; . ./.env; set +a
    python3 backend/db/seed_demo.py --wipe            # API at http://localhost:8080
    python3 backend/db/seed_demo.py --wipe --minimal  # only admin + 1 seller + 2 fixed-price + 2 auctions

Needs: python3, psql, DATABASE_URL in the environment.
All demo accounts use the password in DEMO_PASSWORD.
"""

import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime, timedelta, timezone

API = os.environ.get("API_URL", "http://localhost:8080")
DEMO_PASSWORD = "DemoPass123"


def sql(statement):
    subprocess.run(
        ["psql", os.environ["DATABASE_URL"], "-v", "ON_ERROR_STOP=1", "-qtA", "-c", statement],
        check=True,
    )


def call(method, path, token=None, body=None, want=None):
    req = urllib.request.Request(
        API + path,
        method=method,
        data=None if body is None else json.dumps(body).encode(),
        headers={
            **({"Content-Type": "application/json"} if body is not None else {}),
            **({"Authorization": "Bearer " + token} if token else {}),
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as res:
            status, raw = res.status, res.read().decode()
    except urllib.error.HTTPError as err:
        status, raw = err.code, err.read().decode()
    if want is not None and status != want:
        sys.exit(f"FAILED {method} {path}: expected {want}, got {status}: {raw.strip()}")
    return json.loads(raw) if raw.strip().startswith(("{", "[")) else raw


def account(username, full_name, email):
    call("POST", "/api/register", body={"username": username, "password": DEMO_PASSWORD,
                                         "full_name": full_name, "email": email}, want=201)
    return call("POST", "/api/login", body={"username": username, "password": DEMO_PASSWORD}, want=200)["token"]


def main():
    if "--wipe" not in sys.argv:
        sys.exit("Refusing to run: this deletes ALL data. Re-run with --wipe.")

    print("1. Wiping every table (schema kept, IDs restart at 1)")
    sql("""DO $$ BEGIN
        EXECUTE (SELECT 'TRUNCATE ' || string_agg(format('public.%I', tablename), ', ') || ' RESTART IDENTITY CASCADE'
                 FROM pg_tables WHERE schemaname = 'public');
    END $$;""")

    print("2. Accounts")
    admin = account("admin", "GemBid Admin", "admin@gembid.test")
    sql("UPDATE public.users SET role = 'ADMIN' WHERE username = 'admin'")
    siam = account("siam_gems", "Somchai Rattanakul", "somchai@siamgems.test")
    if "--minimal" in sys.argv:
        minimal(admin, siam)
        return
    chan = account("chanthaburi_jewels", "Malee Srisuk", "malee@chanjewels.test")
    lanna = account("lanna_stones", "Arthit Kaewmanee", "arthit@lannastones.test")
    alice = account("alice", "Alice Wong", "alice@example.test")
    bob = account("bob", "Bob Thompson", "bob@example.test")
    chai = account("chai", "Chaiwat Boonmee", "chai@example.test")

    print("3. Sellers: two approved, one pending")
    for token, shop, legal, kind, addr in [
        (siam, "Siam Gem House", "Siam Gem House Co., Ltd.", "Business", "88 Silom Rd, Bang Rak, Bangkok"),
        (chan, "Chanthaburi Jewels", "Malee Srisuk", "Individual", "Gem Market, Chanthaburi"),
        (lanna, "Lanna Stones", "Arthit Kaewmanee", "Individual", "Nimman Rd, Chiang Mai"),
    ]:
        call("POST", "/api/seller/apply", token, {"shop_name": shop, "legal_name": legal, "seller_type": kind,
                                                   "address": addr, "description": f"{shop}: certified natural gemstones."}, want=201)
    for app in call("GET", "/api/admin/seller-applications", admin, want=200)["applications"]:
        if app["shop_name"] != "Lanna Stones":
            call("PUT", f"/api/admin/seller-applications/{app['seller_id']}", admin, {"decision": "APPROVE"}, want=200)

    # Siam Gem House is PRO (more than 5 live listings); Chanthaburi Jewels stays FREE.
    call("POST", "/api/subscriptions/upgrade", siam, {"payment_method": "MOCK_CARD"}, want=200)

    print("4. Listings")
    now = datetime.now(timezone.utc)

    def listing(token, name, gem_type, desc, carat, price=None, auction=None, draft=False):
        gem = call("POST", "/api/gems", token, {"gem_name": name, "gem_type": gem_type, "description": desc,
                                                 "carat": carat}, want=201)["gem"]
        lid = call("POST", "/api/listings", token, {"gem_id": gem["gem_id"]}, want=201)["listing"]["listing_id"]
        if draft:
            return lid, None
        if price is not None:
            call("PUT", f"/api/listings/{lid}/publish", token, {"sale_type": "FIXED_PRICE", "fixed_price": price}, want=200)
            return lid, None
        start_price, increment, reserve, duration = auction
        call("PUT", f"/api/listings/{lid}/publish", token, {"sale_type": "AUCTION"}, want=200)
        a = call("POST", "/api/auctions", token, {
            "listing_id": lid, "starting_price": start_price, "minimum_increment": increment,
            "reserve_price": reserve, "start_time": now.isoformat(),
            "end_time": (now + duration).isoformat()}, want=201)
        return lid, a["auction"]["auction_id"]

    sapphire, _ = listing(siam, "Royal Blue Sapphire", "Sapphire", "Kanchanaburi blue sapphire, heated, oval cut.", 2.1, price=12500)
    diamond, _ = listing(siam, "Diamond Solitaire Ring", "Diamond", "VS1 round brilliant set in 18k white gold.", 0.8, price=45000)
    emerald, _ = listing(siam, "Emerald Drop Earrings", "Emerald", "Pair of Colombian emeralds, minor oil.", 1.4, price=8500)
    yellow, _ = listing(siam, "Yellow Sapphire", "Sapphire", "Sri Lankan yellow sapphire, unheated.", 3.0, price=6800)
    pearl, _ = listing(siam, "South Sea Pearl Ring", "Pearl", "Golden South Sea pearl, 11 mm.", 4.2, price=3200)
    ruby, ruby_auction = listing(siam, "Pigeon Blood Ruby", "Ruby", "Burmese ruby, unheated, with lab report.", 1.5,
                                 auction=(5000, 250, 6000, timedelta(days=1)))
    listing(siam, "Amethyst Pendant", "Amethyst", "Deep purple Brazilian amethyst on a silver chain.", 6.0,
            auction=(2000, 100, None, timedelta(days=3)))
    listing(siam, "Opal Cabochon", "Other", "Australian black opal, strong play of colour.", 2.6, draft=True)
    topaz, topaz_auction = listing(siam, "Blue Topaz", "Other", "Swiss blue topaz, emerald cut.", 5.0,
                                   auction=(1500, 100, None, timedelta(seconds=75)))
    spinel, spinel_auction = listing(siam, "Red Spinel", "Other", "Mogok red spinel.", 1.2,
                                     auction=(4000, 200, 9000, timedelta(seconds=75)))

    star, _ = listing(chan, "Star Sapphire", "Sapphire", "Black star sapphire from Chanthaburi, sharp six-ray star.", 5.5, price=18000)
    ruby_ring, _ = listing(chan, "Ruby Ring", "Ruby", "Thai ruby in a 14k gold setting.", 1.1, price=9900)
    tsavorite, _ = listing(chan, "Tsavorite Garnet", "Other", "Kenyan tsavorite, vivid green.", 1.8, price=7500)
    garnet, _ = listing(chan, "Rhodolite Garnet", "Other", "Purplish-red rhodolite, cushion cut.", 3.3, price=2500)
    _, chan_auction = listing(chan, "Colombian Emerald", "Emerald", "Muzo emerald, minor oil, with certificate.", 2.0,
                              auction=(7000, 500, None, timedelta(days=2)))

    print("5. Bids")
    for token, auction_id, amount in [(alice, ruby_auction, 5000), (bob, ruby_auction, 5250), (alice, ruby_auction, 5500),
                                      (chai, chan_auction, 7000), (bob, chan_auction, 7500),
                                      (chai, topaz_auction, 1500), (alice, topaz_auction, 1600),
                                      (bob, spinel_auction, 4000)]:
        call("POST", f"/api/auctions/{auction_id}/bids", token, {"bid_amount": amount}, want=201)

    print("6. Orders in every status")
    def buy(token, lid):
        return call("POST", f"/api/listings/{lid}/buy", token, want=201)["order_id"]

    def pay(token, oid, address):
        call("POST", f"/api/orders/{oid}/pay", token, {"payment_method": "MOCK_QR", "shipping_address": address}, want=200)

    oid = buy(alice, emerald)  # DELIVERED
    pay(alice, oid, "99 Sukhumvit 21, Watthana, Bangkok 10110")
    call("PUT", f"/api/orders/{oid}/ship", siam, {"courier": "Kerry Express", "tracking_number": "KEX0013579TH"}, want=200)
    call("POST", f"/api/orders/{oid}/confirm-delivery", alice, want=200)
    oid = buy(bob, yellow)  # SHIPPED
    pay(bob, oid, "12 Nimman Soi 7, Chiang Mai 50200")
    call("PUT", f"/api/orders/{oid}/ship", siam, {"courier": "Thailand Post EMS", "tracking_number": "EF123456789TH"}, want=200)
    pay(chai, buy(chai, pearl), "45 Beach Rd, Pattaya, Chonburi 20150")  # PAID, waiting for shipping
    buy(bob, ruby_ring)  # PENDING_PAYMENT
    call("POST", f"/api/orders/{buy(alice, garnet)}/cancel", alice, want=200)  # CANCELLED, listing back on sale

    print("7. A pending report")
    call("POST", "/api/reports", alice, {"listing_id": tsavorite, "reason": "Suspected synthetic stone",
                                         "description": "The photos and price look too good for natural tsavorite."}, want=201)

    print("8. Waiting for the two short auctions to close themselves (up to 2.5 min)")
    deadline = time.time() + 150
    while time.time() < deadline:
        statuses = [call("GET", f"/api/auctions/{a}", want=200)["status"] for a in (topaz_auction, spinel_auction)]
        if statuses == ["ENDED", "ENDED"]:
            break
        time.sleep(10)
    else:
        print("   (not closed yet; the backend closes them within 30s of ending)")

    stats = call("GET", "/api/admin/analytics", admin, want=200)
    print("Done:", json.dumps({k: stats[k] for k in ("listings_by_status", "auctions_by_status", "orders_by_status")}))
    print(f"All accounts use the password {DEMO_PASSWORD!r}.")


def minimal(admin, seller):
    """Admin + one approved seller with 2 fixed-price listings and 2 auctions (no bids or orders)."""
    call("POST", "/api/seller/apply", seller, {"shop_name": "Siam Gem House", "legal_name": "Siam Gem House Co., Ltd.",
         "seller_type": "Business", "address": "88 Silom Rd, Bang Rak, Bangkok",
         "description": "Siam Gem House: certified natural gemstones."}, want=201)
    app = call("GET", "/api/admin/seller-applications", admin, want=200)["applications"][0]
    call("PUT", f"/api/admin/seller-applications/{app['seller_id']}", admin, {"decision": "APPROVE"}, want=200)

    now = datetime.now(timezone.utc)
    for name, gem_type, desc, carat, price, auction in [
        ("Royal Blue Sapphire", "Sapphire", "Kanchanaburi blue sapphire, heated, oval cut.", 2.1, 12500, None),
        ("Diamond Solitaire Ring", "Diamond", "VS1 round brilliant set in 18k white gold.", 0.8, 45000, None),
        ("Pigeon Blood Ruby", "Ruby", "Burmese ruby, unheated, with lab report.", 1.5, None, (5000, 250, 6000)),
        ("Colombian Emerald", "Emerald", "Muzo emerald, minor oil, with certificate.", 2.0, None, (7000, 500, None)),
    ]:
        gem = call("POST", "/api/gems", seller, {"gem_name": name, "gem_type": gem_type, "description": desc, "carat": carat}, want=201)["gem"]
        lid = call("POST", "/api/listings", seller, {"gem_id": gem["gem_id"]}, want=201)["listing"]["listing_id"]
        if price:
            call("PUT", f"/api/listings/{lid}/publish", seller, {"sale_type": "FIXED_PRICE", "fixed_price": price}, want=200)
        else:
            call("PUT", f"/api/listings/{lid}/publish", seller, {"sale_type": "AUCTION"}, want=200)
            call("POST", "/api/auctions", seller, {"listing_id": lid, "starting_price": auction[0], "minimum_increment": auction[1],
                 "reserve_price": auction[2], "start_time": now.isoformat(), "end_time": (now + timedelta(days=7)).isoformat()}, want=201)
    print(f"Done: admin + siam_gems, 2 fixed-price listings, 2 auctions (7 days). Password {DEMO_PASSWORD!r}.")


if __name__ == "__main__":
    main()
