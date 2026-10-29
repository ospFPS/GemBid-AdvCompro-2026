"use client";

import { useState } from "react";
import { api, getToken } from "./lib/api";

const REASONS = [
	"Counterfeit or fake gemstone",
	"Misleading description or photos",
	"Prohibited or illegal item",
	"Suspected scam or fraud",
	"Inappropriate content",
	"Other",
];

// Report a listing or an auction. Pass exactly one of listingId / auctionId.
export default function ReportButton({ listingId, auctionId }) {
	const [open, setOpen] = useState(false);
	const [reason, setReason] = useState(REASONS[0]);
	const [details, setDetails] = useState("");
	const [busy, setBusy] = useState(false);
	const [done, setDone] = useState(false);
	const [error, setError] = useState("");

	async function submit(event) {
		event.preventDefault();
		if (!getToken()) {
			window.location.href = "/login";
			return;
		}
		setBusy(true);
		setError("");
		try {
			await api("/api/reports", {
				method: "POST",
				body: {
					reason,
					description: details.trim() || undefined,
					...(auctionId ? { auction_id: Number(auctionId) } : { listing_id: Number(listingId) }),
				},
			});
			setDone(true);
		} catch (err) {
			setError(err.message);
		}
		setBusy(false);
	}

	if (done) {
		return <p role="status" className="report-done">Thanks, your report was sent to the moderators.</p>;
	}

	if (!open) {
		return (
			<button type="button" className="report-button" onClick={() => setOpen(true)}>
				⚑ Report this listing
			</button>
		);
	}

	return (
		<form className="report-form" onSubmit={submit}>
			<label htmlFor="report-reason">Reason</label>
			<select id="report-reason" value={reason} onChange={(e) => setReason(e.target.value)}>
				{REASONS.map((r) => <option key={r} value={r}>{r}</option>)}
			</select>
			<label htmlFor="report-details">Details (optional)</label>
			<textarea
				id="report-details"
				rows={3}
				maxLength={2000}
				value={details}
				onChange={(e) => setDetails(e.target.value)}
				placeholder="Anything the moderators should know"
			/>
			{error && <p role="alert">{error}</p>}
			<div className="report-actions">
				<button type="submit" disabled={busy}>{busy ? "Sending..." : "Send report"}</button>
				<button type="button" onClick={() => setOpen(false)} disabled={busy}>Cancel</button>
			</div>
		</form>
	);
}
