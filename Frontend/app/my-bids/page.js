"use client";
import { useState } from "react";
const bids = [
    {
        id: 1,
        gemstone: "Blue Sapphire",
        image: "/assets/Blue sapphire.webp",
        yourBid: "$500",
        currentBid: "$520",
        status: "Won",
        auctionEnds: "2 hours 15 minutes",
    },
    {
        id: 2,
        gemstone: "Ruby Necklace",
        image: "/assets/Ruby necklace.jpg",
        yourBid: "$280",
        currentBid: "$280",
        status: "Winning",
        auctionEnds: "5 hours 30 minutes",
    },
    {
        id: 3,
        gemstone: "Emerald Earrings",
        image: "/assets/Emerald earrings.webp",
        yourBid: "$350",
        currentBid: "$400",
        status: "Lost",
        auctionEnds: "Ended",
    },
    {
        id: 4,
        gemstone: "Diamond Ring",
        image: "/assets/Diamond ring.webp",
        yourBid: "$1000",
        currentBid: "$1200",
        status: "Outbid",
        auctionEnds: "1 hour 45 minutes",
    },
    {
        id: 5,
        gemstone: "Sapphire Earrings",
        image: "/assets/Sapphire earrings.webp",
        yourBid: "$650",
        currentBid: "$650",
        status: "Winning",
        auctionEnds: "3 hours 10 minutes",
    }
];
export default function MyBids() {
    const [filter, setFilter] = useState("active");
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
            <a href="/" className="my-bids-home-link">
                ← Back to Home
            </a>

            <h1>My Bids</h1>
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
                        <button type="button">
                            {bid.status === "Won" ? "View Purchase" : "View Auction"}
                        </button>
                    </section>
                ))}
            </div>
        </main>
    );
}