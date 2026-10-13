export const CART_STORAGE_KEY = "gembid-cart";
export const CART_UPDATED_EVENT = "gembid-cart-updated";
const CART_STORAGE_ERROR = "__gembid_cart_storage_error__";

export function getCartSnapshot() {
	try {
		return window.localStorage.getItem(CART_STORAGE_KEY);
	} catch {
		return CART_STORAGE_ERROR;
	}
}

export function getServerCartSnapshot() {
	return null;
}

export function subscribeToCart(onChange) {
	function handleStorageChange(event) {
		if (event.key === null || event.key === CART_STORAGE_KEY) {
			onChange();
		}
	}

	window.addEventListener(CART_UPDATED_EVENT, onChange);
	window.addEventListener("storage", handleStorageChange);

	return () => {
		window.removeEventListener(CART_UPDATED_EVENT, onChange);
		window.removeEventListener("storage", handleStorageChange);
	};
}

export function parseCartSnapshot(snapshot) {
	if (snapshot === CART_STORAGE_ERROR) {
		return { ids: [], error: "Cart storage is unavailable in this browser." };
	}

	if (snapshot === null) {
		return { ids: [], error: "" };
	}

	try {
		const cartIds = JSON.parse(snapshot);
		if (!Array.isArray(cartIds) || !cartIds.every((id) => typeof id === "string")) {
			throw new Error("The saved cart data is invalid.");
		}

		return { ids: cartIds, error: "" };
	} catch {
		return { ids: [], error: "The saved cart data is invalid." };
	}
}

export function readCartIds() {
	const { ids, error } = parseCartSnapshot(getCartSnapshot());
	if (error) {
		throw new Error(error);
	}

	return ids;
}

export function saveCartIds(cartIds) {
	window.localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(cartIds));
	window.dispatchEvent(new Event(CART_UPDATED_EVENT));
}

export function addCartItem(itemId) {
	saveCartIds([...readCartIds(), itemId]);
}

export function removeCartItem(itemId) {
	const cartIds = readCartIds();
	const itemIndex = cartIds.indexOf(itemId);
	if (itemIndex !== -1) {
		cartIds.splice(itemIndex, 1);
		saveCartIds(cartIds);
	}
}
