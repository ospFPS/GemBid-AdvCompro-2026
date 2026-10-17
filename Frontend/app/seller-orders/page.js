"use client";
import { useEffect, useState } from "react";
import { api, toPurchase } from "../lib/api";

const filters = ["All", "Awaiting Payment", "Processing", "Shipped", "Delivered"];

const PAYMENT_LABELS = { PENDING_PAYMENT: "Pending", CANCELLED: "Cancelled" };

// Seller order + gem + shipping, in the shape this page renders.
async function toSellerOrder(order) {
    const [purchase, detail] = await Promise.all([toPurchase(order), api(`/api/orders/${order.order_id}`)]);
    return {
        id: order.order_id,
        gemstone: purchase.gemstone,
        image: purchase.image,
        buyer: `Buyer #${order.buyer_id}`,
        orderDate: purchase.purchaseDate,
        total: order.total_amount,
        paymentStatus: PAYMENT_LABELS[order.status] ?? "Paid",
        status: purchase.status,
        rawStatus: order.status,
        shippingAddress: detail.shipping?.shipping_address ?? "",
        shippingCarrier: detail.shipping?.courier ?? "",
        trackingNumber: detail.shipping?.tracking_number ?? "",
    };
}

export default function Orders() {
    const [orders, setOrders] = useState([]);
    const [selectedFilter, setSelectedFilter] = useState("All");
    const [messages, setMessages] = useState({});
    const [error, setError] = useState("");

    useEffect(() => {
        api("/api/orders?as=seller")
            .then((list) => Promise.all(list.map(toSellerOrder)))
            .then(setOrders)
            .catch((err) => setError(err.message));
    }, []);

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

    async function saveShipping(orderId) {
        const order = orders.find((item) => item.id === orderId);
        const say = (text) => setMessages((current) => ({ ...current, [orderId]: text }));

        if (!order.shippingCarrier.trim() || !order.trackingNumber.trim()) {
            say("Please enter both the carrier and tracking number.");
            return;
        }
        try {
            await api(`/api/orders/${orderId}/ship`, {
                method: "PUT",
                body: { courier: order.shippingCarrier, tracking_number: order.trackingNumber },
            });
            updateOrder(orderId, "status", "Shipped");
            updateOrder(orderId, "rawStatus", "SHIPPED");
            say("Marked as shipped. The buyer has been notified.");
        } catch (err) {
            say(err.message);
        }
    }

    return (
        <main className="orders-page">
            <a href="/store" className="seller-orders-home">
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

            {error && <p role="alert">{error}</p>}
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

                                    <p>
                                        <strong>Order Status:</strong> {order.status}
                                    </p>
                                    {order.shippingAddress && (
                                        <p>
                                            <strong>Ship to:</strong> {order.shippingAddress}
                                        </p>
                                    )}
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
                                    disabled={order.rawStatus !== "PAID"}
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
                                    disabled={order.rawStatus !== "PAID"}
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
                                    disabled={order.rawStatus !== "PAID"}
                                >
                                    {order.rawStatus === "PENDING_PAYMENT"
                                        ? "Waiting for buyer payment"
                                        : "Mark as Shipped"}
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
