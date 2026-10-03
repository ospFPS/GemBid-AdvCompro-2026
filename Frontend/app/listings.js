// Prototype item
const mockItems = [
	{
		id: "leek-1",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
		price: 12.99,
	},
	{
		id: "amog-1",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 24.5,
	},
	{
		id: "fire-1",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 125.3,
	},
	{
		id: "leek-2",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
		price: 18,
	},
	{
		id: "amog-2",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 31.75,
	},
	{
		id: "fire-2",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 132.2,
	},
	{
		id: "leek-3",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
		price: 15.25,
	},
	{
		id: "amog-3",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 42,
	},
	{
		id: "fire-3",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 532.1,
	},
		{
		id: "leek-4",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
		price: 16.35,
	},
	{
		id: "amog-4",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
		price: 69.69,
	},
	{
		id: "fire-4",
		title: "fire (This is a test item)",
		description: "A gem from teyvat.",
		image: "/assets/smirk.png",
		price: 1253.25,
	},
];

// load listing
export const PAGE_SIZE_OPTIONS = [3, 10, 50];

export async function getListings({
	page = 1,
	pageSize = 20,
	minPrice,
	maxPrice,
} = {}) {
	const filteredItems = mockItems.filter((item) =>
		(minPrice === undefined || item.price >= minPrice)
		&& (maxPrice === undefined || item.price <= maxPrice)
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