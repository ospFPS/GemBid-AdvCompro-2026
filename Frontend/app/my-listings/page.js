"use client";
import { useEffect, useState } from "react";
import { api, toMyListing } from "../lib/api";
export default function MyListings() {
    const [filter, setFilter] = useState("all");
    const [listings, setListings] = useState([]);
    const [error, setError] = useState("");

    useEffect(() => {
        api("/api/seller/listings")
            .then((list) => setListings(list.map(toMyListing)))
            .catch((err) => setError(err.message));
    }, []);
    const filteredListings = listings.filter((listing) => {
        if (filter === "all") {
            return true;
        }

        return listing.status.toLowerCase() === filter;
    });
    return (
        <main className="my-listings-page">
            <div className="my-listings-sticky-nav">
                <a href="/store" className="my-listings-home-link">
                    ← Back to Store
                </a>
                <div className="listing-filters">
                    {["all", "active", "draft", "sold", "unsold"].map(
                        (status) => (
                            <button
                                key={status}
                                type="button"
                                className={filter === status ? "active" : ""}
                                onClick={() => setFilter(status)}
                            >
                                {status.charAt(0).toUpperCase() +
                                    status.slice(1)}
                            </button>
                        )
                    )}
                    <a
                        href="/create-listing"
                        className="create-listing-button"
                    >
                        + Create New Listing
                    </a>
                </div>
            </div>
            <div className="my-listings-heading">
                <div>
                    <h1>My Listings</h1>
                    <p>Manage your gemstone listings and sales.</p>
                </div>
            </div>
            {error && <p role="alert">{error}</p>}
            <div className="listings-list">
                {filteredListings.length > 0 ? (
                    filteredListings.map((listing) => (
                        <section
                            className="listing-card"
                            key={listing.id}
                        >
                            <img
                                src={listing.image}
                                alt={listing.gemstone}
                            />
                            <div className="listing-card-details">
                                <h2>{listing.gemstone}</h2>
                                <p>
                                    <strong>Price:</strong> {listing.price}
                                </p>
                                <p>
                                    <strong>Listing type:</strong>{" "}
                                    {listing.listingType}
                                </p>
                                <p className="listing-status">
                                    <strong>Status:</strong>{" "}
                                    <span
                                        className={`listing-status-${listing.status.toLowerCase()}`}
                                    >
                                        {listing.status}
                                    </span>
                                </p>
                                <div className="listing-actions">
                                    <a href={`/my-listings/view-listings/${listing.id}`} className="listing-view-button">
                                        <button>
                                            View
                                        </button>
                                    </a>
                                    {listing.live && (
                                        <a href={`/item?id=${listing.id}`} className="listing-view-button">
                                            <button>
                                                See in Store
                                            </button>
                                        </a>
                                    )}
                                </div>
                            </div>
                        </section>
                    ))
                ) : (
                    <p className="no-listings">
                        No listings found for this status.
                    </p>
                )}
            </div>
        </main>
    );
}