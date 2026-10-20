"use client";

import { useEffect, useState } from "react";
import { api, formatDate, formatTHB } from "../lib/api";

// Live bid box: polls the auction and its bids every 2s, ticks the countdown every second.
// ponytail: polling, switch to SSE/WebSocket if traffic ever makes 2s polls expensive.
export default function BidForm({ auctionId }) {
	const [auction, setAuction] = useState(null);
	const [bids, setBids] = useState([]);
	const [now, setNow] = useState(() => Date.now());
	const [me, setMe] = useState(null);
	const [amount, setAmount] = useState("");
	const [message, setMessage] = useState("");
	const [busy, setBusy] = useState(false);

	useEffect(() => {
		try {
			setMe(JSON.parse(localStorage.getItem("user"))?.user_id ?? null);
		} catch {}
		const json = (res) => (res.ok ? res.json() : null);
		const load = () => {
			fetch(`/api/auctions/${auctionId}`).then(json).then((data) => data && setAuction(data)).catch(() => {});
			fetch(`/api/auctions/${auctionId}/bids`).then(json).then((data) => data && setBids(data)).catch(() => {});
		};
		load();
		const poll = setInterval(load, 2000);
		const tick = setInterval(() => setNow(Date.now()), 1000);
		return () => {
			clearInterval(poll);
			clearInterval(tick);
		};
	}, [auctionId]);

	if (!auction) {
		return <p>Loading auction...</p>;
	}

	// Mirrors the backend rule: first bid >= current price, then current + increment.
	const minimumBid = auction.current_highest_bidder_id
		? auction.current_price + auction.minimum_increment
		: auction.current_price;
	const msLeft = new Date(auction.end_time) - now;
	const isOpen = auction.status === "ACTIVE" && msLeft > 0;
	const leader = auction.current_highest_bidder_id;
	const myStanding = !me || !bids.some((b) => b.bidder_id === me)
		? null
		: leader === me
			? (isOpen ? "You are the highest bidder." : "You won this auction!")
			: (isOpen ? "You have been outbid." : "Auction ended: you did not win.");

	async function placeBid(event) {
		event.preventDefault();
		setBusy(true);
		setMessage("");
		try {
			const result = await api(`/api/auctions/${auctionId}/bids`, {
				method: "POST",
				body: { bid_amount: Number(amount) },
			});
			setAuction({ ...auction, current_price: result.current_price, current_highest_bidder_id: result.bid.bidder_id });
			setMessage("Bid accepted!");
			setAmount("");
		} catch (err) {
			setMessage(err.message);
		}
		setBusy(false);
	}

	return (
		<form onSubmit={placeBid}>
			<p>
				Current price: <strong>{formatTHB(auction.current_price)}</strong>
				<br />
				Ends: {formatDate(auction.end_time)}
				{isOpen && <> ({timeLeft(msLeft)} left)</>}
				<br />
				Status: {isOpen ? auction.status : "ENDED"}
			</p>
			{myStanding && <p role="status"><strong>{myStanding}</strong></p>}
			<label htmlFor="bid-amount">Bid amount (minimum {formatTHB(minimumBid)})</label>
			<div className="bid-amount-field">
				<span aria-hidden="true">฿</span>
				<input
					id="bid-amount"
					name="bid-amount"
					type="number"
					min={minimumBid}
					step="0.01"
					placeholder={minimumBid.toFixed(2)}
					value={amount}
					onChange={(event) => setAmount(event.target.value)}
					disabled={!isOpen || busy}
					required
				/>
			</div>
			<button className="item-action-button" type="submit" disabled={!isOpen || busy}>
				{isOpen ? (busy ? "Placing bid..." : "Place bid") : "Bidding closed"}
			</button>
			{message && <p role="status">{message}</p>}
			<h3>Latest bids</h3>
			{bids.length === 0 ? (
				<p>No bids yet.</p>
			) : (
				<ol className="bid-history">
					{bids.slice(0, 5).map((b) => (
						<li key={b.bid_id}>
							{formatTHB(b.bid_amount)} by {b.bidder_id === me ? "you" : `Bidder #${b.bidder_id}`}, {formatDate(b.bid_time)}
						</li>
					))}
				</ol>
			)}
		</form>
	);
}

function timeLeft(ms) {
	const sec = Math.floor(ms / 1000);
	const d = Math.floor(sec / 86400), h = Math.floor(sec / 3600) % 24, m = Math.floor(sec / 60) % 60, s = sec % 60;
	return d ? `${d}d ${h}h ${m}m` : h ? `${h}h ${m}m ${s}s` : `${m}m ${s}s`;
}
