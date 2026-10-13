const mockItems = [
	{
		id: "leek-1",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves. Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah Blah",
		image: "/assets/Leek.jpg",
		price: 12.99,
		isBid: false,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "amog-1",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 24.5,
		isBid: true,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "fire-1",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 125.3,
		isBid: false,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "leek-2",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
		price: 18,
		isBid: true,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "amog-2",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 31.75,
		isBid: false,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "fire-2",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 132.2,
		isBid: true,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "leek-3",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
		price: 15.25,
		isBid: false,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "amog-3",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 42,
		isBid: true,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "fire-3",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 532.1,
		isBid: false,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
		{
		id: "leek-4",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
		price: 16.35,
		isBid: true,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "amog-4",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 69.69,
		isBid: false,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
	{
		id: "fire-4",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 1253.25,
		isBid: true,
		pfp: "/assets/pfp.png",
		seller: "Garfield",
	},
];

// load listing
export const PAGE_SIZE_OPTIONS = [3, 10, 50];

export async function getListings({
	page = 1,
	pageSize = 10,
	minPrice,
	maxPrice,
	searchTitle = "",
	listingType = "all",
} = {}) {
	const normalizedTitle = searchTitle.trim().toLowerCase();
	// Filter titles without regard to capitalization.
	const filteredItems = mockItems.filter((item) =>
		(minPrice === undefined || item.price >= minPrice)
		&& (maxPrice === undefined || item.price <= maxPrice)
		&& (listingType === "all"
			|| (listingType === "buy-now" && !item.isBid)
			|| (listingType === "bid" && item.isBid))
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
	return mockItems.find((item) => item.id === id) ?? null;
}

export async function getAllListings() {
	return mockItems;
}

export async function getMoreListings(excludedId, limit = 5) {
	return mockItems.filter((item) => item.id !== excludedId).slice(0, limit);
}