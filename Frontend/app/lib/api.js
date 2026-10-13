// Shared helpers for talking to the GemBid backend.
// Browser code calls relative /api/... paths, which next.config.js proxies to BACKEND_URL.
// Server components use serverApi() from ./server.js instead.

export function getToken() {
	return typeof window === "undefined" ? null : localStorage.getItem("token");
}

export function saveLogin(token, user) {
	localStorage.setItem("token", token);
	localStorage.setItem("user", JSON.stringify(user));
}

export function logout() {
	localStorage.removeItem("token");
	localStorage.removeItem("user");
}

// api calls the backend from the browser. Errors from the backend are plain text,
// so they are thrown as Error(message). A 401 clears the login and goes to /login.
export async function api(path, { method = "GET", body } = {}) {
	const token = getToken();
	const res = await fetch(path, {
		method,
		headers: {
			...(body !== undefined ? { "Content-Type": "application/json" } : {}),
			...(token ? { Authorization: `Bearer ${token}` } : {}),
		},
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	if (res.status === 401) {
		logout();
		window.location.href = "/login";
		throw new Error("Please log in.");
	}
	if (!res.ok) {
		throw new Error((await res.text()).trim() || `Request failed (${res.status})`);
	}
	return res.json();
}

export function formatTHB(amount) {
	return new Intl.NumberFormat("th-TH", { style: "currency", currency: "THB" }).format(amount ?? 0);
}

export function formatDate(value) {
	return value
		? new Date(value).toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" })
		: "-";
}

// The backend has no image upload yet, so show a photo that matches the gem type.
const GEM_IMAGES = {
	sapphire: "/assets/my-bids/Blue sapphire.webp",
	ruby: "/assets/my-bids/Ruby necklace.jpg",
	emerald: "/assets/my-bids/Emerald earrings.webp",
	diamond: "/assets/my-bids/Diamond ring.webp",
	amethyst: "/assets/my-bids/Amethyst.png",
	pearl: "/assets/my-purchases/Pearl ring.webp",
};

// Shrink a picked photo in the browser (max 1000px JPEG) and return it as a data URL for upload.
export function resizeImage(file, maxSize = 1000) {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.onload = () => {
			const scale = Math.min(1, maxSize / Math.max(img.width, img.height));
			const canvas = document.createElement("canvas");
			canvas.width = Math.round(img.width * scale);
			canvas.height = Math.round(img.height * scale);
			canvas.getContext("2d").drawImage(img, 0, 0, canvas.width, canvas.height);
			URL.revokeObjectURL(img.src);
			resolve(canvas.toDataURL("image/jpeg", 0.85));
		};
		img.onerror = () => reject(new Error(`${file.name} is not an image the browser can read.`));
		img.src = URL.createObjectURL(file);
	});
}

export function gemImage(gemType) {
	return GEM_IMAGES[(gemType ?? "").trim().toLowerCase()] ?? "/assets/my-purchases/Yellow sapphire.webp";
}

const ORDER_STATUS_LABELS = {
	PENDING_PAYMENT: "Awaiting Payment",
	PAID: "Processing",
	SHIPPED: "Shipped",
	DELIVERED: "Delivered",
	CANCELLED: "Cancelled",
};

export function orderStatusLabel(status) {
	return ORDER_STATUS_LABELS[status] ?? status;
}

// Orders only carry listing_id, so look up the gem behind a listing (cached per page load).
const gemCache = new Map();

export function gemForListing(listingId) {
	if (!gemCache.has(listingId)) {
		gemCache.set(
			listingId,
			fetch(`/api/listings/${listingId}`)
				.then((res) => (res.ok ? res.json() : null))
				.then((listing) => (listing ? fetch(`/api/gems/${listing.gem_id}`) : null))
				.then((res) => (res?.ok ? res.json() : null))
				.catch(() => null)
		);
	}
	return gemCache.get(listingId);
}

// Order + its gem, in the display shape the purchase pages use.
export async function toPurchase(order) {
	const gem = await gemForListing(order.listing_id);
	return {
		id: order.order_id,
		gemstone: gem?.gem_name ?? `Listing #${order.listing_id}`,
		gemType: gem?.gem_type,
		image: gem?.image_url || gemImage(gem?.gem_type),
		orderId: `ORD-${String(order.order_id).padStart(3, "0")}`,
		purchaseType: order.auction_id ? "Auction" : "Fixed Price",
		price: formatTHB(order.total_amount),
		purchaseDate: formatDate(order.created_at),
		status: orderStatusLabel(order.status),
		rawStatus: order.status,
	};
}

// Seller listing (GET /api/seller/listings) in the shape the My Listings pages use.
// PUBLISHED counts as "Active" for sellers: it is live in the store.
const LISTING_STATUS_LABELS = { PUBLISHED: "Active", ACTIVE: "Active", DRAFT: "Draft", SOLD: "Sold", UNSOLD: "Unsold" };

export function toMyListing(l) {
	const isAuction = l.sale_type === "AUCTION";
	const price = isAuction ? l.current_price : l.fixed_price;
	return {
		id: l.listing_id,
		gemstone: l.gem_name,
		image: l.image_url || gemImage(l.gem_type),
		price: price == null ? "Not set" : formatTHB(price),
		listingType: isAuction ? "Auction" : l.sale_type ? "Fixed Price" : "Not chosen",
		status: LISTING_STATUS_LABELS[l.status] ?? l.status.charAt(0) + l.status.slice(1).toLowerCase(),
		description: l.description ?? "",
		gemstoneType: l.gem_type,
		minimumIncrement: l.minimum_increment == null ? "" : formatTHB(l.minimum_increment),
		endTime: l.end_time,
		live: l.status === "PUBLISHED" || l.status === "ACTIVE",
	};
}
