"use client";
import { useEffect, useState } from "react";
import { api, toPurchase } from "../lib/api";
export default function MyPurchases() {
    const [filter, setFilter] = useState("all");
    const [purchases, setPurchases] = useState([]);
    const [error, setError] = useState("");

    useEffect(() => {
        api("/api/orders")
            .then((orders) => Promise.all(orders.map(toPurchase)))
            .then(setPurchases)
            .catch((err) => setError(err.message));
    }, []);
    const filteredPurchases = purchases.filter((purchase) => {
        if (filter === "fixed") {
            return purchase.purchaseType === "Fixed Price";
        }
        if (filter === "auction") {
            return purchase.purchaseType === "Auction";
        }
        return true;
    });
    return (
        <main className="my-purchases-page">
            <div className="my-purchases-sticky-nav">
                <a href="/store" className="my-purchases-home-link">
                    ← Back to Store
                </a>
                <div className="purchase-filters">
                    <button
                        type="button"
                        className={filter === "all" ? "active" : ""}
                        onClick={() => setFilter("all")}
                    >
                        All
                    </button>
                    <button
                        type="button"
                        className={filter === "fixed" ? "active" : ""}
                        onClick={() => setFilter("fixed")}
                    >
                        Fixed Price
                    </button>
                    <button
                        type="button"
                        className={filter === "auction" ? "active" : ""}
                        onClick={() => setFilter("auction")}
                    >
                        Auction
                    </button>
                </div>
            </div>
            <h1>My Purchases</h1>

            {error && <p role="alert">{error}</p>}
            {!error && purchases.length === 0 && <p>No purchases yet.</p>}
            <div className="purchases-list">
                {filteredPurchases.map((purchase) => (
                    <section
                        className="purchase-card"
                        key={purchase.id}
                    >
                        <h2>{purchase.gemstone}</h2>

                        <img
                            src={purchase.image}
                            alt={purchase.gemstone}
                        />
                        <p>
                            <strong>Order ID:</strong> {purchase.orderId}
                        </p>
                        <p>
                            <strong>Purchase type:</strong>{" "}
                            {purchase.purchaseType}
                        </p>
                        <p>
                            <strong>Price:</strong> {purchase.price}
                        </p>
                        <p>
                            <strong>Purchase date:</strong>{" "}
                            {purchase.purchaseDate}
                        </p>
                        <p
                            className={`purchase-status purchase-status-${purchase.status.toLowerCase().replaceAll(" ", "-")}`}
                        >
                            <strong>Status:</strong> {purchase.status}
                        </p>
                        <a href={`/my-purchases/view-order/${purchase.id}?from=my-purchases`}>
                            <button type="button">View Order</button>
                        </a>
                    </section>
                ))}
            </div>
        </main>
    );
}