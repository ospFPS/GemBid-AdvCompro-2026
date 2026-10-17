"use client";
import { useParams } from "next/navigation";
import { useState } from "react";
import { listings } from "../../../data/listings";

export default function EditListing() {
    const params = useParams();
    const listing = listings.find((item) => item.id === Number(params.id));

    const [gemstoneName, setGemstoneName] = useState(listing?.gemstone || "");
    const [gemstoneType, setGemstoneType] = useState(listing?.gemstoneType || "Sapphire");
    const [description, setDescription] = useState(listing?.description || "");
    const [saleMethod, setSaleMethod] = useState(listing?.listingType || "Fixed Price");
    const [price, setPrice] = useState(listing?.price?.replace("THB ", "") || "");
    const [minimumIncrement, setMinimumIncrement] = useState(listing?.minimumIncrement?.replace("THB ", "") || "");
    const [reservePrice, setReservePrice] = useState(listing?.reservePrice?.replace("THB ", "") || "");
    const [message, setMessage] = useState("");

    if (!listing || !["Draft", "Unsold"].includes(listing.status)) {
        return (
            <main className="edit-listing-page">
                <a href="/my-listings" className="edit-listing-back-link">
                    ← Back to My Listings
                </a>
                <h1>Draft Not Found</h1>
                <p>This listing does not exist or is not a draft that can be edited.</p>
            </main>
        );
    }

    function handleSave(event) {
        event.preventDefault();
        if (listing.status === "Draft") {
            setMessage(
                "Draft changes are ready, but saving edits permanently is not connected yet."
            );
        } else {
            setMessage(
                "Your changes are ready to be relisted, but updating the listing is not connected yet."
            );
        }
    }

    return (
        <main className="edit-listing-page">
            <a href="/my-listings" className="edit-listing-back-link">
                ← Back to My Listings
            </a>         
            <div className="edit-listing-heading">
                <h1>
                    {listing.status === "Draft"
                        ? "Edit Draft"
                        : "Edit Unsold Listing"}
                </h1>
                <p>
                    {listing.status === "Draft"
                        ? "Update your saved listing before publishing it."
                        : "Update your listing details before offering the item again."}
                </p>
            </div>

            <form className="edit-listing-form" onSubmit={handleSave}>
                <section className="edit-listing-section">
                    <h2>Item Details</h2>

                    <div className="edit-listing-group">
                        <label htmlFor="gemstoneName">Listing Name</label>
                        <input
                            id="gemstoneName"
                            type="text"
                            value={gemstoneName}
                            onChange={(event) => setGemstoneName(event.target.value)}
                            required
                        />
                    </div>

                    <div className="edit-listing-group">
                        <label htmlFor="gemstoneType">Gemstone Type</label>
                        <select
                            id="gemstoneType"
                            value={gemstoneType}
                            onChange={(event) => setGemstoneType(event.target.value)}
                        >
                            <option value="Sapphire">Sapphire</option>
                            <option value="Ruby">Ruby</option>
                            <option value="Emerald">Emerald</option>
                            <option value="Diamond">Diamond</option>
                            <option value="Other">Other</option>
                        </select>
                    </div>

                    <div className="edit-listing-group">
                        <label htmlFor="description">Description</label>
                        <textarea
                            id="description"
                            value={description}
                            onChange={(event) => setDescription(event.target.value)}
                            rows="5"
                            required
                        />
                    </div>
                </section>

                <section className="edit-listing-section">
                    <h2>Sale Details</h2>

                    <div className="edit-listing-group">
                        <label htmlFor="saleMethod">Sale Method</label>
                        <select
                            id="saleMethod"
                            value={saleMethod}
                            onChange={(event) => setSaleMethod(event.target.value)}
                        >
                            <option value="Fixed Price">Fixed Price</option>
                            <option value="Auction">Auction</option>
                        </select>
                    </div>

                    <div className="edit-listing-group">
                        <label htmlFor="price">
                            {saleMethod === "Auction" ? "Starting Price (THB)" : "Selling Price (THB)"}
                        </label>
                        <input
                            id="price"
                            type="number"
                            min="1"
                            step="0.01"
                            value={price}
                            onChange={(event) => setPrice(event.target.value)}
                            required
                        />
                    </div>

                    {saleMethod === "Auction" && (
                        <>
                            <div className="edit-listing-group">
                                <label htmlFor="minimumIncrement">Minimum Bid Increment (THB)</label>
                                <input
                                    id="minimumIncrement"
                                    type="number"
                                    min="0.01"
                                    step="0.01"
                                    value={minimumIncrement}
                                    onChange={(event) => setMinimumIncrement(event.target.value)}
                                    required
                                />
                            </div>

                            <div className="edit-listing-group">
                                <label htmlFor="reservePrice">Reserve Price (THB, Optional)</label>
                                <input
                                    id="reservePrice"
                                    type="number"
                                    min="1"
                                    step="0.01"
                                    value={reservePrice}
                                    onChange={(event) => setReservePrice(event.target.value)}
                                />
                            </div>
                        </>
                    )}
                </section>

                {message && (
                    <p className="edit-listing-message" role="status">
                        {message}
                    </p>
                )}

                <div className="edit-listing-actions">
                    {listing.status === "Draft" && (
                        <button type="submit">Save Changes</button>
                    )}
                    {listing.status === "Unsold" && (
                        <button type="submit">
                            Save Changes &amp; Relist
                        </button>
                    )}
                </div>
            </form>
        </main>
    );
}
