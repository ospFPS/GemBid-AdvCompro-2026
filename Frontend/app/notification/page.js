"use client";
import { useState } from "react";
const notifications = [
    {
        id: 1,
        type: "outbid",
        title: "You have been outbid",
        message: "Someone placed a higher bid on Diamond Ring.",
        details: "Current bid: $1200",
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
        message: "Your Ruby ring has been shipped.",
        details: "Order ID: ORD-002",
        time: "Yesterday",
        unread: false,
        link: "/my-purchases/view-order/2?from=notification",
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
        message: "Your Amethyst Earrings purchase has been confirmed.",
        details: "Order total: $2500",
        time: "3 days ago",
        unread: false,
        link: "/my-purchases/view-order/1?from=notification",
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
            <div className="notifications-sticky-nav">
                <a href="/" className="notifications-home-link">
                    ← Back to Home
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
            
            <div className="notifications-list">
                {filteredNotifications.length === 0 ? (
                    <p className="notification-message">No unread messages...</p>
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
                                View Order
                            </a>
                        )}
                        <div>
                            <button
                                type="button"
                                onClick={() => deleteNotification(notification.id)}
                            >
                                Delete
                            </button>
                        </div>
                    </section>
                    ))
                )}
            </div>
        </main>
    );
}