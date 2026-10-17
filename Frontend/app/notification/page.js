"use client";
import { useEffect, useState } from "react";
import { api, formatDate } from "../lib/api";

// Where a notification should take the user.
const SELLER_ORDER_TYPES = ["ORDER_CREATED", "PAYMENT_RECEIVED", "ORDER_DELIVERED", "ORDER_CANCELLED", "AUCTION_SOLD"];

function linkFor(n) {
    if (n.order_id && SELLER_ORDER_TYPES.includes(n.type)) return "/seller-orders";
    if (n.order_id) return `/my-purchases/view-order/${n.order_id}?from=notification`;
    if (n.auction_id && n.listing_id) return `/bid?itemId=${n.listing_id}`;
    if (n.type === "SUBSCRIPTION_ACTIVATED") return "/profile";
    return null;
}

function toCard(n) {
    return {
        id: n.notification_id,
        type: n.type,
        title: n.title,
        message: n.message,
        details: "",
        time: formatDate(n.created_at),
        unread: !n.is_read,
        link: linkFor(n),
    };
}

export default function Notifications() {
    const [notificationList, setNotificationList] = useState([]);
    const [filter, setFilter] = useState("all");
    const [error, setError] = useState("");

    useEffect(() => {
        api("/api/notifications")
            .then((data) => setNotificationList(data.notifications.map(toCard)))
            .catch((err) => setError(err.message));
    }, []);

    const markAllAsRead = () => {
        api("/api/notifications/read-all", { method: "PUT" })
            .then(() => setNotificationList((list) => list.map((n) => ({ ...n, unread: false }))))
            .catch((err) => setError(err.message));
    };
    const markAsRead = (id) => {
        api(`/api/notifications/${id}/read`, { method: "PUT" })
            .then(() => setNotificationList((list) => list.map((n) => (n.id === id ? { ...n, unread: false } : n))))
            .catch((err) => setError(err.message));
    };
    const filteredNotifications = notificationList.filter((notification) => {
        if (filter === "unread") {
            return notification.unread;
        }
        if (filter === "read") {
            return !notification.unread;
        }
        return true;
    });
    return (
        <main className="notifications-page">
            <div className="notifications-sticky-nav">
                <a href="/store" className="notifications-home-link">
                    ← Back to Store
                </a>
                <div className="notification-filters">
                    <button
                        type="button"
                        className={filter === "all" ? "active" : ""}
                        onClick={() => setFilter("all")}
                    >
                        All
                    </button>
                    <button
                        type="button"
                        className={filter === "unread" ? "active" : ""}
                        onClick={() => setFilter("unread")}
                    >
                        Unread
                    </button>
                    <button
                        type="button"
                        className={filter === "read" ? "active" : ""}
                        onClick={() => setFilter("read")}
                    >
                        Read
                    </button>
                </div>
                <button
                    className="mark-all-read-button"
                    type="button"
                    onClick={markAllAsRead}
                >
                    Mark all as read
                </button>
            </div>
                <h1>Notifications</h1>
            {error && <p role="alert">{error}</p>}
            
            <div className="notifications-list">
                {filteredNotifications.length === 0 ? (
                    <p className="notification-message">No notifications here.</p>
                ) : (
                    filteredNotifications.map((notification) => (
                    <section
                        className="notification-card"
                        key={notification.id}
                    >
                        <h2>{notification.title}</h2>
                        <p>{notification.message}</p>
                        <p>{notification.details}</p>
                        <p className="notification-time">
                            {notification.time}
                        </p>
                        {notification.link && (
                            <a href={notification.link}>
                                View
                            </a>
                        )}
                        {notification.unread && (
                            <div>
                                <button
                                    type="button"
                                    onClick={() => markAsRead(notification.id)}
                                >
                                    Mark as read
                                </button>
                            </div>
                        )}
                    </section>
                    ))
                )}
            </div>
        </main>
    );
}