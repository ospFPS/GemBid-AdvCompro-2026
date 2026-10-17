"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { api, formatDate, formatTHB, gemImage } from "../lib/api";

// One row per auction from GET /api/me/bids.
function toBidCard(b) {
    const ended = b.auction_status === "ENDED";
    return {
        id: b.auction_id,
        listingId: b.listing_id,
        gemstone: b.gem_name,
        image: b.image_url || gemImage(b.gem_type),
        yourBid: formatTHB(b.my_highest_bid),
        currentBid: formatTHB(b.current_price),
        status: ended ? (b.is_winner ? "Won" : "Lost") : b.is_leading ? "Winning" : "Outbid",
        auctionEnds: ended ? "Ended" : formatDate(b.end_time),
    };
}

export default function MyBids() {
    const [filter, setFilter] = useState("active");
    const [bids, setBids] = useState([]);
    const [error, setError] = useState("");

    useEffect(() => {
        api("/api/me/bids")
            .then((list) => setBids(list.map(toBidCard)))
            .catch((err) => setError(err.message));
    }, []);
    const filteredBids = bids.filter((bid) => {
        if (filter === "active") {
            return bid.status === "Winning" || bid.status === "Outbid";
        }
        if (filter === "won") {
            return bid.status === "Won";
        }
        if (filter === "lost") {
            return bid.status === "Lost";
        }
        return true;
    });
    return (
        <main className="my-bids-page">
            <div className="my-bids-sticky-nav">
                <a href="/store" className="my-bids-home-link">
                    ← Back to Store
                </a>
                <div className="bid-filters">
                    <button
                        type="button"
                        className={filter === "active" ? "active" : ""}
                        onClick={() => setFilter("active")}
                    >
                        Active
                    </button>
                    <button
                        type="button"
                        className={filter === "won" ? "active" : ""}
                        onClick={() => setFilter("won")}
                    >
                        Won
                    </button>
                    <button
                        type="button"
                        className={filter === "lost" ? "active" : ""}
                        onClick={() => setFilter("lost")}
                    >
                        Lost
                    </button>
                </div>
            </div>
            <h1>My Bids</h1>
            {error && <p role="alert">{error}</p>}
            {!error && filteredBids.length === 0 && <p>No bids here yet.</p>}
            
            <div className="bids-list">
                {filteredBids.map((bid) => (
                    <section
                        className="bid-card"
                        key={bid.id}
                    >
                        <h2>{bid.gemstone}</h2>
                        <img
                            src={bid.image}
                            alt={bid.gemstone}
                        />
                        <p>
                            <strong>Your bid:</strong> {bid.yourBid}
                        </p>
                        <p>
                            <strong>Current bid:</strong> {bid.currentBid}
                        </p>
                        <p className={`bid-status bid-status-${bid.status.toLowerCase()}`}>
                            <strong>Status:</strong> {bid.status}
                        </p>
                        <p>
                            <strong>Auction ends:</strong> {bid.auctionEnds}
                        </p>
                        <Link
                            href={
                            bid.status === "Won"
                                ? `/my-purchases`
                                : `/bid?itemId=${bid.listingId}`
                            }
                            className="bid-action-button"
                        >
                            {bid.status === "Won" ? "View Purchase" : "View Auction"}
                        </Link>
                    </section>
                ))}
            </div>
        </main>
    );
}