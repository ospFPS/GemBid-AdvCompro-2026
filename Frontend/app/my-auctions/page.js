"use client";
import { useState } from "react";

const auctions = [
    {
        id: 1,
        gemstone: "Blue Sapphire",
        image: "/assets/my-bids/Blue sapphire.webp",
        startingPrice: 5000,
        currentBid: 6200,
        bidCount: 12,
        endTime: "12 October 2026, 10:00 PM",
        status: "Active",
    },
    {
        id: 2,
        gemstone: "Ruby Necklace",
        image: "/assets/my-bids/Ruby necklace.jpg",
        startingPrice: 3000,
        currentBid: 3000,
        bidCount: 0,
        endTime: "15 October 2026, 8:00 PM",
        status: "Upcoming",
    },
    {
        id: 3,
        gemstone: "Emerald Earrings",
        image: "/assets/my-bids/Emerald earrings.webp",
        startingPrice: 4000,
        currentBid: 5500,
        bidCount: 8,
        endTime: "10 October 2026, 6:00 PM",
        status: "Ended",
    },
    {
        id: 4,
        gemstone: "Yellow Sapphire",
        image: "/assets/my-bids/Sapphire earrings.webp",
        startingPrice: 5000,
        currentBid: 7500,
        bidCount: 15,
        endTime: "9 October 2026, 8:00 PM",
        status: "Sold",
    },
    {
        id: 5,
        gemstone: "Amethyst",
        image: "/assets/my-bids/Amethyst.png",
        startingPrice: 2000,
        currentBid: 2000,
        bidCount: 0,
        endTime: "8 October 2026, 5:00 PM",
        status: "Unsold",
    },
];

const filters = [
    "All",
    "Upcoming",
    "Active",
    "Ended",
    "Sold",
    "Unsold",
];

export default function MyAuctions() {
    const [selectedFilter, setSelectedFilter] = useState("All");

    const filteredAuctions =
        selectedFilter === "All"
            ? auctions
            : auctions.filter(
                (auction) => auction.status === selectedFilter
            );

    return (
        <main className="my-auctions-page">
            <a href="/" className="my-auctions-home-link">
                ← Back to Home
            </a>
            <div className="my-auctions-heading">
                <h1>My Auctions</h1>
                <p>Monitor and manage your gemstone auctions.</p>
            </div>

            <div className="my-auctions-filters">
                {filters.map((filter) => (
                    <button
                        key={filter}
                        type="button"
                        className={
                            selectedFilter === filter ? "active" : ""
                        }
                        onClick={() => setSelectedFilter(filter)}
                    >
                        {filter}
                    </button>
                ))}
            </div>

            <div className="my-auctions-grid">
                {filteredAuctions.length > 0 ? (
                    filteredAuctions.map((auction) => (
                        <article
                            className="my-auction-card"
                            key={auction.id}
                        >
                            <img
                                src={auction.image}
                                alt={auction.gemstone}
                            />

                            <div className="my-auction-details">
                                <div className="my-auction-title-row">
                                    <h2>{auction.gemstone}</h2>
                                    <span
                                        className={`my-auction-status status-${auction.status.toLowerCase()}`}
                                    >
                                        {auction.status}
                                    </span>
                                </div>

                                <p>
                                    Starting Price: THB{" "}
                                    {auction.startingPrice.toLocaleString()}
                                </p>

                                <p>
                                    Current Highest Bid: THB{" "}
                                    {auction.currentBid.toLocaleString()}
                                </p>

                                <p>
                                    Total Bids: {auction.bidCount}
                                </p>

                                <p>
                                    {auction.status === "Upcoming"
                                        ? "Starts:"
                                        : "Ends:"}{" "}
                                    {auction.endTime}
                                </p>

                                <a href="/store" className="my-auction-view-link">
                                    View Auction
                                </a>
                            </div>
                        </article>
                    ))
                ) : (
                    <p className="my-auctions-empty">
                        No auctions found in this category.
                    </p>
                )}
            </div>
        </main>
    );
}
