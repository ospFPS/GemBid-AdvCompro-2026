
"use client";
import { useState } from "react";

const salesData = {
    "7 Days": [
        { label: "Oct 6", sales: 800 },
        { label: "Oct 7", sales: 1200 },
        { label: "Oct 8", sales: 1500 },
        { label: "Oct 9", sales: 2800 },
        { label: "Oct 10", sales: 2200 },
        { label: "Oct 11", sales: 1000 },
        { label: "Oct 12", sales: 1500 },
    ],
    "30 Days": [
        { label: "Week 1", sales: 6500 },
        { label: "Week 2", sales: 8200 },
        { label: "Week 3", sales: 5400 },
        { label: "Week 4", sales: 7100 },
    ],
    "90 Days": [
        { label: "August", sales: 18000 },
        { label: "September", sales: 25000 },
        { label: "October", sales: 11000 },
    ],
};

const topGemstones = [
    { name: "Yellow Sapphire", sold: 1, revenue: 7500 },
    { name: "Emerald Earrings", sold: 1, revenue: 3500 },
    { name: "Ruby Necklace", sold: 0, revenue: 0 },
];

const auctionStats = [
    { status: "Active", count: 1 },
    { status: "Upcoming", count: 1 },
    { status: "Ended", count: 1 },
    { status: "Sold", count: 1 },
    { status: "Unsold", count: 1 },
];

const orderStats = [
    { status: "Processing", count: 2 },
    { status: "Shipped", count: 1 },
    { status: "Delivered", count: 0 },
];

export default function SellerAnalytics() {
    const [selectedRange, setSelectedRange] = useState("7 Days");

    const currentSales = salesData[selectedRange];
    const highestSales = Math.max(
        ...currentSales.map((item) => item.sales),
        1
    );

    const totalSales = 11000;
    const itemsSold = 2;
    const activeAuctions = 1;
    const pendingOrders = 2;

    return (
        <main className="seller-analytics-page">
            <a href="/" className="analytics-home-link">
                ← Back to Home
            </a>

            <div className="analytics-heading">
                <h1>Seller Analytics</h1>
                <p>
                    Track your sales, gemstone performance, auctions,
                    and customer orders.
                </p>
            </div>

            <p className="analytics-demo-notice">
                Demo data: These figures are examples and are not connected
                to actual sales or orders.
            </p>

            {/* Summary Cards */}
            <section className="analytics-summary">
                <article className="analytics-summary-card">
                    <h2>Total Sales</h2>
                    <p className="analytics-summary-value">
                        THB {totalSales.toLocaleString()}
                    </p>
                    <p className="analytics-summary-note">
                        Revenue from sold items
                    </p>
                </article>

                <article className="analytics-summary-card">
                    <h2>Items Sold</h2>
                    <p className="analytics-summary-value">{itemsSold}</p>
                    <p className="analytics-summary-note">
                        Successfully sold items
                    </p>
                </article>

                <article className="analytics-summary-card">
                    <h2>Active Auctions</h2>
                    <p className="analytics-summary-value">
                        {activeAuctions}
                    </p>
                    <p className="analytics-summary-note">
                        Auctions currently running
                    </p>
                </article>

                <article className="analytics-summary-card">
                    <h2>Pending Orders</h2>
                    <p className="analytics-summary-value">
                        {pendingOrders}
                    </p>
                    <p className="analytics-summary-note">
                        Orders being processed
                    </p>
                </article>
            </section>

            {/* Sales Chart */}
            <section className="analytics-section">
                <div className="analytics-section-heading">
                    <div>
                        <h2>Sales Overview</h2>
                        <p>Review your sales over time.</p>
                    </div>

                    <select
                        aria-label="Sales date range"
                        value={selectedRange}
                        onChange={(event) =>
                            setSelectedRange(event.target.value)
                        }
                    >
                        <option value="7 Days">Last 7 Days</option>
                        <option value="30 Days">Last 30 Days</option>
                        <option value="90 Days">Last 90 Days</option>
                    </select>
                </div>

                <div className="analytics-chart">
                    {currentSales.map((item) => (
                        <div className="analytics-chart-column" key={item.label}>
                            <span className="analytics-chart-value">
                                {item.sales.toLocaleString()}
                            </span>

                            <div className="analytics-chart-bar-area">
                                <div
                                    className="analytics-chart-bar"
                                    style={{
                                        height: `${Math.max(
                                            (item.sales / highestSales) * 100,
                                            3
                                        )}%`,
                                    }}
                                    title={`THB ${item.sales.toLocaleString()}`}
                                />
                            </div>

                            <span className="analytics-chart-label">
                                {item.label}
                            </span>
                        </div>
                    ))}
                </div>

                <p className="analytics-chart-caption">
                    Sales amount (THB). Chart values are illustrative.
                </p>
            </section>

            {/* Top-Selling Gemstones */}
            <section className="analytics-section">
                <div className="analytics-section-heading">
                    <div>
                        <h2>Top-Selling Gemstones</h2>
                        <p>Compare the performance of your items.</p>
                    </div>
                </div>

                <div className="analytics-table-wrapper">
                    <table className="analytics-table">
                        <thead>
                            <tr>
                                <th>Gemstone</th>
                                <th>Items Sold</th>
                                <th>Total Sales</th>
                            </tr>
                        </thead>

                        <tbody>
                            {topGemstones.map((gemstone) => (
                                <tr key={gemstone.name}>
                                    <td>{gemstone.name}</td>
                                    <td>{gemstone.sold}</td>
                                    <td>
                                        THB {gemstone.revenue.toLocaleString()}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </section>

            {/* Auction Performance */}
            <section className="analytics-section">
                <div className="analytics-section-heading">
                    <div>
                        <h2>Auction Performance</h2>
                        <p>See the current status of your auctions.</p>
                    </div>
                </div>

                <div className="analytics-stat-list">
                    {auctionStats.map((auction) => (
                        <div
                            className="analytics-stat-row"
                            key={auction.status}
                        >
                            <span>{auction.status}</span>
                            <strong>{auction.count}</strong>
                        </div>
                    ))}
                </div>
            </section>

            {/* Order Statistics */}
            <section className="analytics-section">
                <div className="analytics-section-heading">
                    <div>
                        <h2>Order Statistics</h2>
                        <p>Monitor the fulfilment status of customer orders.</p>
                    </div>
                </div>

                <div className="analytics-stat-list">
                    {orderStats.map((order) => (
                        <div
                            className="analytics-stat-row"
                            key={order.status}
                        >
                            <span>{order.status}</span>
                            <strong>{order.count}</strong>
                        </div>
                    ))}
                </div>

                <a href="/seller-orders" className="analytics-orders-link">
                    View Orders →
                </a>
            </section>
        </main>
    );
}
