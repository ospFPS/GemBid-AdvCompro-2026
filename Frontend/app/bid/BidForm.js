"use client";

import { useEffect, useState } from "react";
import { api, formatDate, formatTHB } from "../lib/api";

// Live bid box: polls the auction every 5s and places real bids.
export default function BidForm({ auctionId }) {
	const [auction, setAuction] = useState(null);
	const [amount, setAmount] = useState("");
	const [message, setMessage] = useState("");
	const [busy, setBusy] = useState(false);

	useEffect(() => {
		const load = () =>
			fetch(`/api/auctions/${auctionId}`)
				.then((res) => (res.ok ? res.json() : null))
				.then((data) => data && setAuction(data))
				.catch(() => {});
		load();
		const timer = setInterval(load, 5000);
		return () => clearInterval(timer);
	}, [auctionId]);

	if (!auction) {
		return <p>Loading auction...</p>;
	}

	// Mirrors the backend rule: first bid >= current price, then current + increment.
	const minimumBid = auction.current_highest_bidder_id
		? auction.current_price + auction.minimum_increment
		: auction.current_price;
	const isOpen = auction.status === "ACTIVE" && new Date(auction.end_time) > new Date();

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
			setMessage(`Bid accepted! Next minimum bid: ${formatTHB(result.next_minimum_bid)}`);
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
				<br />
				Status: {auction.status}
			</p>
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
		</form>
	);
}
