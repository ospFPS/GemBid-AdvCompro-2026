"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getToken, logout, saveLogin } from "./lib/api";

// Account links for the top nav. Shows Log in / Sign up when logged out.
export default function AccountNav() {
	const [user, setUser] = useState(null);

	useEffect(() => {
		const token = getToken();
		if (!token) return;
		// Refresh the stored user so seller status is current (e.g. after admin approval).
		fetch("/api/me", { headers: { Authorization: `Bearer ${token}` } })
			.then((res) => (res.ok ? res.json() : null))
			.then((me) => {
				if (me) {
					saveLogin(token, me);
					setUser(me);
				} else {
					logout();
				}
			})
			.catch(() => {});
	}, []);

	if (!user) {
		return (
			<ul>
				<li><Link className="nav-link" href="/">Return to Home</Link></li>
				<li><Link className="nav-link" href="/login">Log in</Link></li>
				<li><Link className="nav-link" href="/signup">Sign up</Link></li>
			</ul>
		);
	}

	const isSeller = user.seller_status === "VERIFIED";
	return (
		<ul>
			<li><span className="nav-link">Hi, {user.username}</span></li>
			{user.role === "ADMIN" && (
				<li><Link className="nav-link" href="/admin">Admin Dashboard</Link></li>
			)}
			<li><Link className="nav-link" href="/cart">Cart</Link></li>
			<li><Link className="nav-link" href="/my-purchases">My Purchases</Link></li>
			<li><Link className="nav-link" href="/my-bids">My Bids</Link></li>
			<li><Link className="nav-link" href="/notification">Notifications</Link></li>
			{isSeller ? (
				<>
					<li><Link className="nav-link" href="/create-listing">Create Listing</Link></li>
					<li><Link className="nav-link" href="/my-listings">My Listings</Link></li>
					<li><Link className="nav-link" href="/seller-orders">Seller Orders</Link></li>
				</>
			) : (
				<li><Link className="nav-link" href="/become-seller">Become a Seller</Link></li>
			)}
			<li><Link className="nav-link" href="/profile">Profile</Link></li>
			<li>
				<button
					type="button"
					className="nav-link"
					onClick={() => {
						logout();
						window.location.href = "/store";
					}}
				>
					Log out
				</button>
			</li>
		</ul>
	);
}
