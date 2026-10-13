import { serverApi } from "./lib/server";
import { gemImage } from "./lib/api";

// Map a backend listing (GET /api/listings) to the item shape the pages use.
function toItem(listing) {
	const isBid = listing.sale_type === "AUCTION";
	return {
		id: String(listing.listing_id),
		title: listing.gem_name,
		description: listing.description || `${listing.gem_type} gemstone`,
		image: gemImage(listing.gem_type),
		price: (isBid ? listing.current_price : listing.fixed_price) ?? 0,
		isBid,
		pfp: "/assets/pfp.png",
		seller: listing.shop_name,
		gemType: listing.gem_type,
		auctionId: listing.auction_id,
		minimumIncrement: listing.minimum_increment,
		endTime: listing.end_time,
	};
}

async function fetchItems() {
	return ((await serverApi("/api/listings")) ?? []).map(toItem);
}

// load listing
export const PAGE_SIZE_OPTIONS = [3, 10, 50];

export async function getListings({
	page = 1,
	pageSize = 10,
	minPrice,
	maxPrice,
	searchTitle = "",
} = {}) {
	const normalizedTitle = searchTitle.trim().toLowerCase();
	// Filter titles without regard to capitalization.
	const filteredItems = (await fetchItems()).filter((item) =>
		(minPrice === undefined || item.price >= minPrice)
		&& (maxPrice === undefined || item.price <= maxPrice)
		&& item.title.toLowerCase().includes(normalizedTitle)
	);
	const total = filteredItems.length;
	const totalPages = Math.max(1, Math.ceil(total / pageSize));
	const currentPage = Math.min(Math.max(page, 1), totalPages);
	const start = (currentPage - 1) * pageSize;

	return {
		items: filteredItems.slice(start, start + pageSize),
		page: currentPage,
		pageSize,
		total,
		totalPages,
	};
}

export async function getListingById(id) {
	return (await fetchItems()).find((item) => item.id === String(id)) ?? null;
}

export async function getAllListings() {
	return fetchItems();
}

export async function getMoreListings(excludedId, limit = 5) {
	return (await fetchItems()).filter((item) => item.id !== excludedId).slice(0, limit);
}