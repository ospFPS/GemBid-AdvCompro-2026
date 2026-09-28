const itemsContainer = document.querySelector("#items");

function loadItems() {
    const items = [
        {
            "title": "Leek (This is a test item)",
            "description": "A rare gem from the northern caves.",
            "image": "../assets/Leek.jpg"
        },
        {
            "title": "Amog (This is a test item)",
            "description": "A gem from mars.",
            "image": "../assets/Reaction.jpg"
        },
                {
            "title": "Leek (This is a test item)",
            "description": "A rare gem from the northern caves.",
            "image": "../assets/Leek.jpg"
        },
        {
            "title": "Amog (This is a test item)",
            "description": "A gem from mars.",
            "image": "../assets/Reaction.jpg"
        },
                {
            "title": "Leek (This is a test item)",
            "description": "A rare gem from the northern caves.",
            "image": "../assets/Leek.jpg"
        },
        {
            "title": "Amog (This is a test item)",
            "description": "A gem from mars.",
            "image": "../assets/Reaction.jpg"
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