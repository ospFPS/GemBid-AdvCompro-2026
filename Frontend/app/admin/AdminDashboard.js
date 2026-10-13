"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { getAllListings } from "../listings";

const STORAGE_KEY = "gembid.admin.listings";
const EMPTY_FORM = {
	title: "",
	description: "",
	image: "",
	price: "",
	seller: "",
	isBid: false,
};

function validateListings(value) {
	if (!Array.isArray(value)) {
		throw new Error("The file must contain a JSON array of listings.");
	}

	const listings = value.map((listing, index) => {
		if (
			!listing
			|| typeof listing.id !== "string"
			|| typeof listing.title !== "string"
			|| typeof listing.description !== "string"
			|| typeof listing.image !== "string"
			|| typeof listing.seller !== "string"
			|| !listing.title.trim()
			|| !listing.description.trim()
			|| !listing.image.trim()
			|| !listing.seller.trim()
			|| typeof listing.price !== "number"
			|| !Number.isFinite(listing.price)
			|| listing.price < 0
			|| typeof listing.isBid !== "boolean"
		) {
			throw new Error(`Listing ${index + 1} has missing or invalid fields.`);
		}

		return {
			...listing,
			pfp: typeof listing.pfp === "string" ? listing.pfp : "/assets/pfp.png",
			isActive: listing.isActive !== false,
		};
	});

	if (new Set(listings.map((listing) => listing.id)).size !== listings.length) {
		throw new Error("Every listing must have a unique ID.");
	}

	return listings;
}

function formatPrice(price) {
	return new Intl.NumberFormat("en-US", {
		style: "currency",
		currency: "USD",
	}).format(price);
}

export default function AdminDashboard() {
	const [listings, setListings] = useState([]);
	const [ready, setReady] = useState(false);
	const [search, setSearch] = useState("");
	const [typeFilter, setTypeFilter] = useState("all");
	const [statusFilter, setStatusFilter] = useState("all");
	const [sortOrder, setSortOrder] = useState("title");
	const [form, setForm] = useState(EMPTY_FORM);
	const [editingId, setEditingId] = useState(null);
	const [formOpen, setFormOpen] = useState(false);
	const [error, setError] = useState("");
	const [notice, setNotice] = useState("");
	const importInput = useRef(null);

	useEffect(() => {
		let cancelled = false;

		async function loadListings() {
			try {
				const savedListings = window.localStorage.getItem(STORAGE_KEY);
				const initialListings = savedListings === null
					? await getAllListings()
					: validateListings(JSON.parse(savedListings));
				if (!cancelled) {
					setListings(initialListings.map((listing) => ({
						...listing,
						isActive: listing.isActive !== false,
					})));
				}
			} catch (loadError) {
				if (!cancelled) {
					setError(loadError instanceof Error
						? `Could not load saved listings: ${loadError.message}`
						: "Could not load saved listings.");
				}
			} finally {
				if (!cancelled) setReady(true);
			}
		}

		loadListings();
		return () => {
			cancelled = true;
		};
	}, []);

	function saveListings(nextListings, successMessage) {
		try {
			window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextListings));
			setListings(nextListings);
			setError("");
			setNotice(successMessage);
			return true;
		} catch (saveError) {
			setError(saveError instanceof Error
				? `Could not save listings: ${saveError.message}`
				: "Could not save listings in this browser.");
			return false;
		}
	}

	const activeCount = listings.filter((listing) => listing.isActive !== false).length;
	const bidCount = listings.filter((listing) => listing.isBid && listing.isActive !== false).length;
	const filteredListings = useMemo(() => {
		const normalizedSearch = search.trim().toLowerCase();
		const filtered = listings.filter((listing) => {
			const matchesSearch = `${listing.title} ${listing.seller} ${listing.description}`
				.toLowerCase()
				.includes(normalizedSearch);
			const matchesType = typeFilter === "all"
				|| (typeFilter === "bid" && listing.isBid)
				|| (typeFilter === "buy-now" && !listing.isBid);
			const matchesStatus = statusFilter === "all"
				|| (statusFilter === "active" && listing.isActive !== false)
				|| (statusFilter === "hidden" && listing.isActive === false);
			return matchesSearch && matchesType && matchesStatus;
		});

		return filtered.sort((first, second) => {
			if (sortOrder === "price-low") return first.price - second.price;
			if (sortOrder === "price-high") return second.price - first.price;
			return first.title.localeCompare(second.title);
		});
	}, [listings, search, sortOrder, statusFilter, typeFilter]);

	function openCreateForm() {
		setEditingId(null);
		setForm(EMPTY_FORM);
		setFormOpen(true);
		setError("");
		setNotice("");
	}

	function openEditForm(listing) {
		setEditingId(listing.id);
		setForm({
			title: listing.title,
			description: listing.description,
			image: listing.image,
			price: String(listing.price),
			seller: listing.seller,
			isBid: listing.isBid,
		});
		setFormOpen(true);
		setError("");
		setNotice("");
	}

	function handleSubmit(event) {
		event.preventDefault();
		const price = Number(form.price);
		if (
			!form.title.trim()
			|| !form.description.trim()
			|| !form.image.trim()
			|| !form.seller.trim()
			|| !Number.isFinite(price)
			|| price < 0
		) {
			setError("Enter a title, description, image path, seller, and a valid non-negative price.");
			return;
		}

		const editing = editingId !== null;
		const current = listings.find((listing) => listing.id === editingId);
		const nextListing = {
			id: current?.id ?? `listing-${crypto.randomUUID()}`,
			title: form.title.trim(),
			description: form.description.trim(),
			image: form.image.trim(),
			price,
			isBid: form.isBid,
			pfp: current?.pfp ?? "/assets/pfp.png",
			seller: form.seller.trim(),
			isActive: current?.isActive ?? true,
		};
		const nextListings = editing
			? listings.map((listing) => listing.id === editingId ? nextListing : listing)
			: [nextListing, ...listings];

		if (saveListings(nextListings, editing ? "Listing updated." : "Listing created.")) {
			setFormOpen(false);
		}
	}

	function toggleListing(listing) {
		const nextListings = listings.map((current) => current.id === listing.id
			? { ...current, isActive: current.isActive === false }
			: current);
		saveListings(nextListings, listing.isActive === false
			? "Listing published."
			: "Listing hidden from the store.");
	}

	function deleteListing(listing) {
		if (!window.confirm(`Delete "${listing.title}"? This cannot be undone.`)) return;
		saveListings(
			listings.filter((current) => current.id !== listing.id),
			"Listing deleted.",
		);
	}

	function exportListings() {
		try {
			const blob = new Blob([JSON.stringify(listings, null, 2)], {
				type: "application/json",
			});
			const url = URL.createObjectURL(blob);
			const link = document.createElement("a");
			link.href = url;
			link.download = "gembid-listings.json";
			link.click();
			window.setTimeout(() => URL.revokeObjectURL(url), 1000);
			setNotice("Listings exported.");
			setError("");
		} catch (exportError) {
			setError(exportError instanceof Error
				? `Could not export listings: ${exportError.message}`
				: "Could not export listings.");
		}
	}

	async function importListings(event) {
		const file = event.target.files?.[0];
		event.target.value = "";
		if (!file) return;

		try {
			const imported = validateListings(JSON.parse(await file.text()));
			saveListings(imported, `${imported.length} listings imported.`);
		} catch (importError) {
			setError(importError instanceof Error
				? `Could not import listings: ${importError.message}`
				: "Could not import listings.");
		}
	}

	return (
		<main className="admin-page">
			<header className="admin-header">
				<div>
					<p className="admin-eyebrow">GemBid / Management</p>
					<h1>Admin dashboard</h1>
					<p className="admin-subtitle">Manage marketplace listings and storefront visibility.</p>
				</div>
				<Link className="admin-back-link" href="/store">View storefront</Link>
			</header>

			<p className="admin-demo-notice">
				Frontend-only demo. Changes are saved in this browser; server persistence and administrator sign-in are not connected.
			</p>

			<section className="admin-stats" aria-label="Listing overview">
				<div className="admin-stat">
					<span>Total listings</span>
					<strong>{listings.length}</strong>
				</div>
				<div className="admin-stat">
					<span>Published</span>
					<strong>{activeCount}</strong>
				</div>
				<div className="admin-stat">
					<span>Hidden</span>
					<strong>{listings.length - activeCount}</strong>
				</div>
				<div className="admin-stat">
					<span>Active bids</span>
					<strong>{bidCount}</strong>
				</div>
			</section>

			<section className="admin-listings" aria-labelledby="admin-listings-title">
				<div className="admin-section-heading">
					<div>
						<h2 id="admin-listings-title">Listings</h2>
						<p>{filteredListings.length} of {listings.length} listings</p>
					</div>
					<div className="admin-toolbar">
						<input
							ref={importInput}
							className="admin-file-input"
							type="file"
							accept="application/json,.json"
							onChange={importListings}
							aria-label="Import listings JSON file"
						/>
						<button className="admin-button admin-button--secondary" type="button" onClick={exportListings}>
							Export JSON
						</button>
						<button className="admin-button admin-button--secondary" type="button" onClick={() => importInput.current?.click()}>
							Import JSON
						</button>
						<button className="admin-button admin-button--primary" type="button" onClick={openCreateForm}>
							Add listing
						</button>
					</div>
				</div>

				{formOpen && (
					<form className="admin-form" onSubmit={handleSubmit}>
						<div className="admin-form-heading">
							<h3>{editingId ? "Edit listing" : "New listing"}</h3>
							<button className="admin-text-button" type="button" onClick={() => setFormOpen(false)}>
								Cancel
							</button>
						</div>
						<div className="admin-form-grid">
							<label>
								Title
								<input required maxLength="120" value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} />
							</label>
							<label>
								Seller
								<input required maxLength="80" value={form.seller} onChange={(event) => setForm({ ...form, seller: event.target.value })} />
							</label>
							<label className="admin-form-wide">
								Description
								<textarea required rows="3" maxLength="1000" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
							</label>
							<label>
								Image path or URL
								<input required value={form.image} placeholder="/assets/example.jpg" onChange={(event) => setForm({ ...form, image: event.target.value })} />
							</label>
							<label>
								Price (USD)
								<input required type="number" min="0" step="0.01" value={form.price} onChange={(event) => setForm({ ...form, price: event.target.value })} />
							</label>
							<label className="admin-checkbox-label">
								<input type="checkbox" checked={form.isBid} onChange={(event) => setForm({ ...form, isBid: event.target.checked })} />
								This listing uses bidding
							</label>
						</div>
						<div className="admin-form-actions">
							<button className="admin-button admin-button--primary" type="submit">
								{editingId ? "Save changes" : "Create listing"}
							</button>
							<button className="admin-button admin-button--secondary" type="button" onClick={() => setFormOpen(false)}>
								Cancel
							</button>
						</div>
					</form>
				)}

				<div className="admin-filters">
					<label className="admin-search">
						Search listings
						<input
							type="search"
							value={search}
							onChange={(event) => setSearch(event.target.value)}
							placeholder="Search title, seller, or description"
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
						Visibility
						<select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}>
							<option value="all">All listings</option>
							<option value="active">Published</option>
							<option value="hidden">Hidden</option>
						</select>
					</label>
					<label>
						Sort by
						<select value={sortOrder} onChange={(event) => setSortOrder(event.target.value)}>
							<option value="title">Title</option>
							<option value="price-low">Price: low to high</option>
							<option value="price-high">Price: high to low</option>
						</select>
					</label>
				</div>

				{error && <p className="admin-feedback admin-feedback--error" role="alert">{error}</p>}
				{notice && !error && <p className="admin-feedback" role="status">{notice}</p>}
				{!ready ? (
					<p className="admin-empty-state">Loading listings…</p>
				) : filteredListings.length === 0 ? (
					<p className="admin-empty-state">No listings match these filters.</p>
				) : (
					<div className="admin-table-wrap">
						<table className="admin-table">
							<thead>
								<tr>
									<th scope="col">Listing</th>
									<th scope="col">Type</th>
									<th scope="col">Price</th>
									<th scope="col">Visibility</th>
									<th scope="col">Actions</th>
								</tr>
							</thead>
							<tbody>
								{filteredListings.map((listing) => (
									<tr key={listing.id}>
										<td>
											<div className="admin-listing-cell">
												<img src={listing.image} alt="" />
												<div>
													<strong>{listing.title}</strong>
													<span>{listing.seller}</span>
												</div>
											</div>
										</td>
										<td>{listing.isBid ? "Bidding" : "Buy now"}</td>
										<td>{formatPrice(listing.price)}</td>
										<td>
											<span className={`admin-status ${listing.isActive === false ? "admin-status--hidden" : ""}`}>
												{listing.isActive === false ? "Hidden" : "Published"}
											</span>
										</td>
										<td>
											<div className="admin-row-actions">
												<button className="admin-text-button" type="button" onClick={() => openEditForm(listing)}>Edit</button>
												<button className="admin-text-button" type="button" onClick={() => toggleListing(listing)}>
													{listing.isActive === false ? "Publish" : "Hide"}
												</button>
												<button className="admin-text-button admin-text-button--danger" type="button" onClick={() => deleteListing(listing)}>Delete</button>
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