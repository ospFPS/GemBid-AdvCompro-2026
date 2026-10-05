"use client";
import { useState } from "react";
export default function Profile() {
    const [editing, setEditing] = useState(false);
    return (
        <main className="profile-page">
            <a href="/" className="profile-home-link">
                ← Back to Home
            </a>
            <h1>Profile</h1>
            <section className="profile-card">
                <h2>Account Information</h2>
                {editing ? (
                    <div className="profile-edit-form">
                        <label htmlFor="username">
                            Username
                        </label>
                        <input
                            type="text"
                            id="username"
                            defaultValue="username"
                        />
                        <label htmlFor="name">
                            Full Name
                        </label>
                        <input
                            type="text"
                            id="name"
                            defaultValue="Guest User"
                        />
                        <label htmlFor="email">
                            Email
                        </label>
                        <input
                            type="email"
                            id="email"
                           defaultValue="user@example.com"
                        />
                        <label htmlFor="password">
                            Password
                        </label>
                        <input
                            type="password"
                            id="password"
                            placeholder="Enter new password"
                        />
                        <label htmlFor="phone">
                            Phone Number
                        </label>
                        <input
                            type="tel"
                            id="phone"
                            defaultValue="123-456-7890"
                        />
                        <div className="profile-edit-actions">
                            <button type="button">
                                Save Changes
                            </button>
                            <button
                                type="button"
                                onClick={() => setEditing(false)}
                            >
                                Cancel
                            </button>
                        </div>
                    </div>
                ) : (
                    <div className="profile-information">
                        <p>
                            <strong>Username:</strong> username
                        </p>
                        <p>
                            <strong>Full Name:</strong> Guest User
                        </p>
                        <p>
                            <strong>Email:</strong> user@example.com
                        </p>
                        <p>
                            <strong>Phone Number:</strong> 123-456-7890
                        </p>
                    </div>
                )}
                <p>
                    <strong>Account Type:</strong> Normal User
                </p>
            </section>
            <section className="profile-card">
                <h2>User Status</h2>
                <div className="profile-status">
                <p>
                    You are currently a general user.
                </p>
                </div>
                <a href="/become-seller?from=profile" className="profile-link">
                    Become a Seller
                </a>
            </section>
            <section className="profile-card">
                <h2>Account Settings</h2>
                <div className="profile-actions">
                    <button onClick={() => setEditing(true)}>
                        Edit Profile
                    </button>
                    <button>Log Out</button>
                </div>
            </section>
        </main>
    );
}