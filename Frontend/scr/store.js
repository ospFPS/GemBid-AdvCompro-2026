const itemsContainer = document.querySelector("#items");

function loadItems() {
    const items = [
        {
            "title": "Leek",
            "description": "A rare gem from the northern caves.",
            "image": "../assets/Leek.jpg"
        },
        {
            "title": "Amog",
            "description": "A gem from mars.",
            "image": "../assets/Leek.jpg"
        },
                {
            "title": "Leek",
            "description": "A rare gem from the northern caves.",
            "image": "../assets/Leek.jpg"
        },
        {
            "title": "Amog",
            "description": "A gem from mars.",
            "image": "../assets/Leek.jpg"
        },
                {
            "title": "Leek",
            "description": "A rare gem from the northern caves.",
            "image": "../assets/Leek.jpg"
        },
        {
            "title": "Amog",
            "description": "A gem from mars.",
            "image": "../assets/Leek.jpg"
        },
    ];

    for (const item of items) {
        const card = document.createElement("article");
        const image = document.createElement("img");
        const title = document.createElement("h2");
        const description = document.createElement("p");

        card.className = "store-item";
        image.src = item.image;
        image.alt = item.title;
        title.textContent = item.title;
        description.textContent = item.description;

        card.append(image, title, description);
        itemsContainer.append(card);
    }
}

loadItems();