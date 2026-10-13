"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { api, formatDate, formatTHB, gemImage } from "../lib/api";

const LIVE = ["PUBLISHED", "ACTIVE"];
const SUSPENDABLE = ["DRAFT", "PUBLISHED", "ACTIVE"];
const REPORT_ACTIONS = ["NONE", "WARN_USER", "SUSPEND_LISTING", "SUSPEND_SELLER"];

function listingPrice(l) {
	const price = l.sale_type === "AUCTION" ? l.current_price : l.fixed_price;
	return price == null ? "-" : formatTHB(price);
}

function reportTarget(r) {
	if (r.listing_id) return `Listing #${r.listing_id}`;
	if (r.auction_id) return `Auction #${r.auction_id}`;
	return `User #${r.reported_user_id}`;
}

export default function AdminDashboard() {
	const [access, setAccess] = useState("checking"); // checking | admin | denied
	const [analytics, setAnalytics] = useState(null);
	const [applications, setApplications] = useState([]);
	const [listings, setListings] = useState([]);
	const [reports, setReports] = useState([]);
	const [reportActions, setReportActions] = useState({});
	const [search, setSearch] = useState("");
	const [typeFilter, setTypeFilter] = useState("all");
	const [statusFilter, setStatusFilter] = useState("all");
	const [sortOrder, setSortOrder] = useState("newest");
	const [error, setError] = useState("");
	const [notice, setNotice] = useState("");

	const load = useCallback(async () => {
		const [stats, apps, allListings, pending] = await Promise.all([
			api("/api/admin/analytics"),
			api("/api/admin/seller-applications"),
			api("/api/admin/listings"),
			api("/api/admin/reports?status=PENDING"),
		]);
		setAnalytics(stats);
		setApplications(apps.applications);
		setListings(allListings);
		setReports(pending.reports);
	}, []);

	useEffect(() => {
		api("/api/me")
			.then((me) => {
				if (me.role !== "ADMIN") {
					setAccess("denied");
					return;
				}
				setAccess("admin");
				return load();
			})
			.catch((err) => {
				setAccess("denied");
				setError(err.message);
			});
	}, [load]);

	// Run an admin action, show its result, then reload everything.
	async function act(task, successMessage) {
		setError("");
		setNotice("");
		try {
			await task();
			setNotice(successMessage);
			await load();
		} catch (err) {
			setError(err.message);
		}
	}

	const reviewSeller = (app, decision) =>
		act(
			() => api(`/api/admin/seller-applications/${app.seller_id}`, { method: "PUT", body: { decision } }),
			`${app.shop_name}: ${decision === "APPROVE" ? "approved" : "rejected"}.`
		);

	const suspendListing = (listing) => {
		if (!window.confirm(`Suspend "${listing.gem_name}"? Any running auction is stopped without a winner.`)) return;
		act(() => api(`/api/admin/listings/${listing.listing_id}/suspend`, { method: "PUT" }), `"${listing.gem_name}" suspended.`);
	};

	const reviewReport = (report, decision) => {
		const action = decision === "RESOLVE" ? reportActions[report.report_id] ?? "NONE" : "NONE";
		act(
			() => api(`/api/admin/reports/${report.report_id}`, { method: "PUT", body: { decision, action } }),
			`Report #${report.report_id} ${decision === "RESOLVE" ? "resolved" : "rejected"}.`
		);
	};

	const filteredListings = useMemo(() => {
		const query = search.trim().toLowerCase();
		const filtered = listings.filter((l) =>
			(typeFilter === "all"
				|| (typeFilter === "buy-now" && l.sale_type === "FIXED_PRICE")
				|| (typeFilter === "bid" && l.sale_type === "AUCTION"))
			&& (statusFilter === "all"
				|| (statusFilter === "live" ? LIVE.includes(l.status) : l.status === statusFilter))
			&& `${l.gem_name} ${l.shop_name} ${l.gem_type}`.toLowerCase().includes(query)
		);
		const price = (l) => (l.sale_type === "AUCTION" ? l.current_price : l.fixed_price) ?? 0;
		return filtered.sort((a, b) =>
			sortOrder === "price-low" ? price(a) - price(b)
				: sortOrder === "price-high" ? price(b) - price(a)
					: b.listing_id - a.listing_id
		);
	}, [listings, search, typeFilter, statusFilter, sortOrder]);

	if (access !== "admin") {
		return (
			<main className="admin-page">
				<header className="admin-header">
					<div>
						<p className="admin-eyebrow">GemBid / Management</p>
						<h1>Admin dashboard</h1>
					</div>
					<Link className="admin-back-link" href="/store">View storefront</Link>
				</header>
				<p className="admin-empty-state">
					{access === "checking" ? "Checking access…" : "Admin access required. Log in with an admin account."}
				</p>
				{error && <p className="admin-feedback admin-feedback--error" role="alert">{error}</p>}
			</main>
		);
	}

	const pendingApps = applications.filter((a) => a.verification_status === "PENDING").length;
	const liveCount = listings.filter((l) => LIVE.includes(l.status)).length;

	return (
		<main className="admin-page">
			<header className="admin-header">
				<div>
					<p className="admin-eyebrow">GemBid / Management</p>
					<h1>Admin dashboard</h1>
					<p className="admin-subtitle">Seller applications, listings, reports and platform numbers.</p>
				</div>
				<Link className="admin-back-link" href="/store">View storefront</Link>
			</header>

			{error && <p className="admin-feedback admin-feedback--error" role="alert">{error}</p>}
			{notice && !error && <p className="admin-feedback" role="status">{notice}</p>}

			<section className="admin-stats" aria-label="Platform overview">
				<div className="admin-stat">
					<span>Users</span>
					<strong>{analytics?.users.total ?? "-"}</strong>
				</div>
				<div className="admin-stat">
					<span>Live listings</span>
					<strong>{liveCount}</strong>
				</div>
				<div className="admin-stat">
					<span>Pending sellers</span>
					<strong>{pendingApps}</strong>
				</div>
				<div className="admin-stat">
					<span>Paid orders</span>
					<strong>{analytics?.sales.paid_orders ?? "-"}</strong>
				</div>
				<div className="admin-stat">
					<span>Sales (GMV)</span>
					<strong>{formatTHB(analytics?.sales.gross_merchandise_value)}</strong>
				</div>
				<div className="admin-stat">
					<span>PRO members</span>
					<strong>{analytics?.subscriptions.active_pro ?? "-"}</strong>
				</div>
			</section>

			<section className="admin-listings" aria-labelledby="admin-sellers-title">
				<div className="admin-section-heading">
					<div>
						<h2 id="admin-sellers-title">Seller applications</h2>
						<p>{pendingApps} waiting for review, {applications.length} in total</p>
					</div>
				</div>
				{applications.length === 0 ? (
					<p className="admin-empty-state">No seller applications yet.</p>
				) : (
					<div className="admin-table-wrap">
						<table className="admin-table">
							<thead>
								<tr>
									<th scope="col">Shop</th>
									<th scope="col">Applicant</th>
									<th scope="col">Applied</th>
									<th scope="col">Status</th>
									<th scope="col">Actions</th>
								</tr>
							</thead>
							<tbody>
								{applications.map((app) => (
									<tr key={app.seller_id}>
										<td>
											<div className="admin-listing-cell">
												<div>
													<strong>{app.shop_name}</strong>
													<span>{app.seller_type ?? "-"} · {app.legal_name}</span>
												</div>
											</div>
										</td>
										<td>{app.username}</td>
										<td>{formatDate(app.applied_at)}</td>
										<td>
											<span className={`admin-status ${app.verification_status === "VERIFIED" ? "" : "admin-status--hidden"}`}>
												{app.verification_status}
											</span>
										</td>
										<td>
											{app.verification_status === "PENDING" ? (
												<div className="admin-row-actions">
													<button className="admin-text-button" type="button" onClick={() => reviewSeller(app, "APPROVE")}>Approve</button>
													<button className="admin-text-button admin-text-button--danger" type="button" onClick={() => reviewSeller(app, "REJECT")}>Reject</button>
												</div>
											) : "-"}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}
			</section>

			<section className="admin-listings" aria-labelledby="admin-listings-title">
				<div className="admin-section-heading">
					<div>
						<h2 id="admin-listings-title">Listings</h2>
						<p>{filteredListings.length} of {listings.length} listings</p>
					</div>
				</div>

				<div className="admin-filters">
					<label className="admin-search">
						Search listings
						<input
							type="search"
							value={search}
							onChange={(event) => setSearch(event.target.value)}
							placeholder="Search gem, shop or type"
						/>
					</label>
					<label>
						Type
						<select value={typeFilter} onChange={(event) => setTypeFilter(event.target.value)}>
							<option value="all">All types</option>
							<option value="buy-now">Buy now</option>
							<option value="bid">Bidding</option>
						</select>
					</label>
					<label>
						Status
						<select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}>
							<option value="all">All listings</option>
							<option value="live">Live</option>
							<option value="DRAFT">Draft</option>
							<option value="SOLD">Sold</option>
							<option value="UNSOLD">Unsold</option>
							<option value="SUSPENDED">Suspended</option>
						</select>
					</label>
					<label>
						Sort by
						<select value={sortOrder} onChange={(event) => setSortOrder(event.target.value)}>
							<option value="newest">Newest</option>
							<option value="price-low">Price: low to high</option>
							<option value="price-high">Price: high to low</option>
						</select>
					</label>
				</div>

				{filteredListings.length === 0 ? (
					<p className="admin-empty-state">No listings match these filters.</p>
				) : (
					<div className="admin-table-wrap">
						<table className="admin-table">
							<thead>
								<tr>
									<th scope="col">Listing</th>
									<th scope="col">Type</th>
									<th scope="col">Price</th>
									<th scope="col">Status</th>
									<th scope="col">Actions</th>
								</tr>
							</thead>
							<tbody>
								{filteredListings.map((listing) => (
									<tr key={listing.listing_id}>
										<td>
											<div className="admin-listing-cell">
												<img src={gemImage(listing.gem_type)} alt="" />
												<div>
													<strong>{listing.gem_name}</strong>
													<span>{listing.shop_name}</span>
												</div>
											</div>
										</td>
										<td>{listing.sale_type === "AUCTION" ? "Bidding" : listing.sale_type ? "Buy now" : "-"}</td>
										<td>{listingPrice(listing)}</td>
										<td>
											<span className={`admin-status ${LIVE.includes(listing.status) ? "" : "admin-status--hidden"}`}>
												{listing.status}
											</span>
										</td>
										<td>
											<div className="admin-row-actions">
												{LIVE.includes(listing.status) && (
													<Link className="admin-text-button" href={`/item?id=${listing.listing_id}`}>View</Link>
												)}
												{SUSPENDABLE.includes(listing.status) && (
													<button className="admin-text-button admin-text-button--danger" type="button" onClick={() => suspendListing(listing)}>
														Suspend
													</button>
												)}
											</div>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}
			</section>

			<section className="admin-listings" aria-labelledby="admin-reports-title">
				<div className="admin-section-heading">
					<div>
						<h2 id="admin-reports-title">Pending reports</h2>
						<p>{reports.length} waiting for review</p>
					</div>
				</div>
				{reports.length === 0 ? (
					<p className="admin-empty-state">No pending reports.</p>
				) : (
					<div className="admin-table-wrap">
						<table className="admin-table">
							<thead>
								<tr>
									<th scope="col">Target</th>
									<th scope="col">Reason</th>
									<th scope="col">Reported</th>
									<th scope="col">Action if resolved</th>
									<th scope="col">Decision</th>
								</tr>
							</thead>
							<tbody>
								{reports.map((report) => (
									<tr key={report.report_id}>
										<td>{reportTarget(report)}</td>
										<td>
											<div className="admin-listing-cell">
												<div>
													<strong>{report.reason}</strong>
													<span>{report.description ?? ""}</span>
												</div>
											</div>
										</td>
										<td>{formatDate(report.created_at)}</td>
										<td>
											<select
												value={reportActions[report.report_id] ?? "NONE"}
												onChange={(event) => setReportActions({ ...reportActions, [report.report_id]: event.target.value })}
											>
												{REPORT_ACTIONS.map((a) => (
													<option key={a} value={a}>{a.replaceAll("_", " ").toLowerCase()}</option>
												))}
											</select>
										</td>
										<td>
											<div className="admin-row-actions">
												<button className="admin-text-button" type="button" onClick={() => reviewReport(report, "RESOLVE")}>Resolve</button>
												<button className="admin-text-button admin-text-button--danger" type="button" onClick={() => reviewReport(report, "REJECT")}>Reject</button>
											</div>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}
			</section>
		</main>
	);
}
