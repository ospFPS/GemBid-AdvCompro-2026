"use client";
import { useState } from "react";
const notifications = [
    {
        id: 1,
        type: "outbid",
        title: "You have been outbid",
        message: "Someone placed a higher bid on Ruby Necklace.",
        details: "Current bid: $280",
        time: "10 minutes ago",
        unread: true,
    },
    {
        id: 2,
        type: "auction-won",
        title: "Auction Won",
        message: "Congratulations! You won Blue Sapphire.",
        details: "Winning bid: $520",
        time: "2 hours ago",
        unread: true,
    },
    {
        id: 3,
        type: "shipping",
        title: "Order Shipped",
        message: "Your Ruby Ring has been shipped.",
        details: "Tracking number: RB123456",
        time: "Yesterday",
        unread: false,
    },
    {
        id: 4,
        type: "seller",
        title: "Seller Application Approved",
        message: "Your application to become a seller has been approved.",
        details: "You can now create gemstone listings.",
        time: "2 days ago",
        unread: false,
    },
    {
        id: 5,
        type: "order",
        title: "Purchase Confirmed",
        message: "Your purchase of Emerald Earrings has been confirmed.",
        details: "Order total: $350",
        time: "3 days ago",
        unread: false,
    },
];

export default function Notifications() {
    const [notificationList, setNotificationList] = useState(notifications);
    const [filter, setFilter] = useState("all");
    const markAllAsRead = () => {
        setNotificationList(
            notificationList.map((notification) => ({
                ...notification,
                unread: false,
            }))
        );
    };
    const deleteNotification = (id) => {
        setNotificationList(
            notificationList.filter(
                (notification) => notification.id !== id
            )
        );
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
            <a href="/" className="notifications-home-link">
                ← Back to Home
            </a>
            <h1>Notifications</h1>
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
                type="button"
                onClick={markAllAsRead}
            >
                Mark all as read
            </button>
            <div className="notifications-list">
                {filteredNotifications.length === 0 ? (
                    <p>No unread messages.</p>
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
                        <button
                            type="button"
                            onClick={() => deleteNotification(notification.id)}
                        >
                            Delete
                        </button>
                    </section>
                    ))
                )}
            </div>
        </main>
    );
}