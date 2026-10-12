"use client";

import { useParams } from "next/navigation";
import { listings } from "../../../data/listings";

export default function ViewListing() {
    const params = useParams();
    const listing = listings.find((item) => item.id === Number(params.id));

    if (!listing) {
        return (
            <main className="view-listing-page">
                <a href="/my-listings" className="view-listing-back-link">
                    ← Back to My Listings
                </a>
                <h1>Listing Not Found</h1>
                <p>We couldn't find a listing with this ID.</p>
            </main>
        );
    }

    return (
        <main className="view-listing-page">
            <a href="/" className="view-listings-home">
                <button type="button">Home</button>
            </a>
            <div>
                <a href="/my-listings" className="view-listing-back-link">
                    ← Back to My Listings
                </a>
            </div>
            <div className="view-listing-content">
                <div className="view-listing-image">
                    <img src={listing.image} alt={listing.gemstone} />
                </div>

                <div className="view-listing-details">
                    <h1>{listing.gemstone}</h1>

                    <p className="view-listing-price">{listing.price}</p>

                    <p>
                        <strong>Gemstone Type:</strong> {listing.gemstoneType}
                    </p>

                    <p>
                        <strong>Sale Method:</strong> {listing.listingType}
                    </p>

                    <p>
                        <strong>Status:</strong>{" "}
                        <span className={`view-listing-status view-listing-status-${listing.status.toLowerCase()}`}>
                            {listing.status}
                        </span>
                    </p>

                    <div className="view-listing-description">
                        <h2>Description</h2>
                        <p>{listing.description}</p>
                    </div>

                    {listing.listingType === "Auction" && (
                        <div className="view-listing-auction">
                            <h2>Auction Details</h2>
                            <p>
                                <strong>Starting Price:</strong> {listing.price}
                            </p>
                            <p>
                                <strong>Minimum Bid Increment:</strong>{" "}
                                {listing.minimumIncrement || "Not specified"}
                            </p>
                            <p>
                                <strong>Reserve Price:</strong>{" "}
                                {listing.reservePrice || "Not specified"}
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </main>
    );
}
