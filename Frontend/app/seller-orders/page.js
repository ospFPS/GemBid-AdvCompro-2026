"use client";
import { useState } from "react";

const sampleOrders = [
    {
        id: 1,
        gemstone: "Emerald Earrings",
        image: "/assets/my-bids/Emerald earrings.webp",
        buyer: "Alex Johnson",
        orderDate: "10 October 2026",
        total: 3500,
        paymentStatus: "Paid",
        status: "Processing",
        shippingCarrier: "",
        trackingNumber: "",
    },
    {
        id: 2,
        gemstone: "Ruby Necklace",
        image: "/assets/my-bids/Ruby necklace.jpg",
        buyer: "Emily Chen",
        orderDate: "9 October 2026",
        total: 8500,
        paymentStatus: "Paid",
        status: "Shipped",
        shippingCarrier: "DHL",
        trackingNumber: "DHL123456",
    },
    {
        id: 3,
        gemstone: "Blue Sapphire",
        image: "/assets/my-bids/Blue sapphire.webp",
        buyer: "Daniel Smith",
        orderDate: "8 October 2026",
        total: 6200,
        paymentStatus: "Pending",
        status: "Processing",
        shippingCarrier: "",
        trackingNumber: "",
    },
];

const filters = ["All", "Processing", "Shipped", "Delivered"];

export default function Orders() {
    const [orders, setOrders] = useState(sampleOrders);
    const [selectedFilter, setSelectedFilter] = useState("All");
    const [messages, setMessages] = useState({});

    const filteredOrders =
        selectedFilter === "All"
            ? orders
            : orders.filter((order) => order.status === selectedFilter);

    function updateOrder(orderId, field, value) {
        setOrders((currentOrders) =>
            currentOrders.map((order) =>
                order.id === orderId
                    ? { ...order, [field]: value }
                    : order
            )
        );
    }

    function saveShipping(orderId) {
        const order = orders.find((item) => item.id === orderId);

        if (!order.shippingCarrier.trim() || !order.trackingNumber.trim()) {
            setMessages((current) => ({
                ...current,
                [orderId]: "Please enter both the carrier and tracking number.",
            }));
            return;
        }

        setMessages((current) => ({
            ...current,
            [orderId]: "Shipping information saved for this demo.",
        }));
    }

    return (
        <main className="orders-page">
            <a href="/" className="seller-orders-home">
                <button type="button">Home</button>
            </a>
            <div>
                <a href="/seller-analytics" className="orders-analytics-link">
                    ← Back to Analytics
                </a>
            </div>
            <div className="orders-heading">
                <h1>Orders</h1>
                <p>View customer orders and manage shipping information.</p>
            </div>

            <div className="orders-filters">
                {filters.map((filter) => (
                    <button
                        key={filter}
                        type="button"
                        className={selectedFilter === filter ? "active" : ""}
                        onClick={() => setSelectedFilter(filter)}
                    >
                        {filter}
                    </button>
                ))}
            </div>

            <div className="orders-list">
                {filteredOrders.length > 0 ? (
                    filteredOrders.map((order) => (
                        <article className="order-card" key={order.id}>
                            <div className="order-item">
                                <img src={order.image} alt={order.gemstone} />

                                <div className="order-details">
                                    <h2>{order.gemstone}</h2>
                                    <p>
                                        <strong>Order ID:</strong> #{order.id}
                                    </p>
                                    <p>
                                        <strong>Buyer:</strong> {order.buyer}
                                    </p>
                                    <p>
                                        <strong>Order Date:</strong>{" "}
                                        {order.orderDate}
                                    </p>
                                    <p>
                                        <strong>Total:</strong> THB{" "}
                                        {order.total.toLocaleString()}
                                    </p>
                                    <p>
                                        <strong>Payment:</strong>{" "}
                                        <span
                                            className={`payment-status payment-${order.paymentStatus.toLowerCase()}`}
                                        >
                                            {order.paymentStatus}
                                        </span>
                                    </p>

                                    <label htmlFor={`status-${order.id}`}>
                                        Order Status
                                    </label>
                                    <select
                                        id={`status-${order.id}`}
                                        value={order.status}
                                        onChange={(event) =>
                                            updateOrder(
                                                order.id,
                                                "status",
                                                event.target.value
                                            )
                                        }
                                    >
                                        <option value="Processing">
                                            Processing
                                        </option>
                                        <option value="Shipped">Shipped</option>
                                        <option value="Delivered">
                                            Delivered
                                        </option>
                                    </select>
                                </div>
                            </div>

                            <div className="order-shipping">
                                <h3>Shipping Information</h3>

                                <label htmlFor={`carrier-${order.id}`}>
                                    Shipping Carrier
                                </label>
                                <input
                                    id={`carrier-${order.id}`}
                                    type="text"
                                    placeholder="e.g. DHL, FedEx"
                                    value={order.shippingCarrier}
                                    onChange={(event) =>
                                        updateOrder(
                                            order.id,
                                            "shippingCarrier",
                                            event.target.value
                                        )
                                    }
                                />

                                <label htmlFor={`tracking-${order.id}`}>
                                    Tracking Number
                                </label>
                                <input
                                    id={`tracking-${order.id}`}
                                    type="text"
                                    placeholder="Enter tracking number"
                                    value={order.trackingNumber}
                                    onChange={(event) =>
                                        updateOrder(
                                            order.id,
                                            "trackingNumber",
                                            event.target.value
                                        )
                                    }
                                />

                                <button
                                    type="button"
                                    onClick={() => saveShipping(order.id)}
                                >
                                    Save Shipping Information
                                </button>

                                {messages[order.id] && (
                                    <p className="order-message">
                                        {messages[order.id]}
                                    </p>
                                )}
                            </div>
                        </article>
                    ))
                ) : (
                    <p className="orders-empty">
                        No orders found in this category.
                    </p>
                )}
            </div>
        </main>
    );
}
