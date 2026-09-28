// Prototype item
const mockItems = [
	{
		id: "leek-1",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
	},
	{
		id: "amog-1",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
	},
	{
		id: "leek-2",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
	},
	{
		id: "amog-2",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
	},
	{
		id: "leek-3",
		title: "Leek (This is a test item)",
		description: "A rare gem from the northern caves.",
		image: "/assets/Leek.jpg",
	},
	{
		id: "amog-3",
		title: "Amog (This is a test item)",
		description: "A gem from mars.",
		image: "/assets/Reaction.jpg",
	},
];

// load listing
export const PAGE_SIZE_OPTIONS = [10, 20, 50];

export async function getListings({ page = 1, pageSize = 20 } = {}) {
	const total = mockItems.length;
	const totalPages = Math.max(1, Math.ceil(total / pageSize));
	const currentPage = Math.min(Math.max(page, 1), totalPages);
	const start = (currentPage - 1) * pageSize;

	return {
		items: mockItems.slice(start, start + pageSize),
		page: currentPage,
		pageSize,
		total,
		totalPages,
	};
}

export async function getListingById(id) {
	return mockItems.find((item) => item.id === id) ?? null;
}