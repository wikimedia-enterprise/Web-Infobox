# Wiki Infobox System

A dynamic, component that generates Wikipedia hover-cards for articles and blogs.

This repo is broken down into two branches. Main, contains the current demo of the project, hosted [here](https://wikimedia-enterprise.github.io/Web-Infobox/).

The WIP branch contains the current version of the packaged form of this project, for use by other developers. As time goes on the WIP branch will undergo changes.

## Features
- **Go Backend:** Scrapes and cleans Wikipedia data

- **Themeable Frontend:** A Javascript and CSS library powered by CSS variables, making it easy to adapt to any style without touching the core code.

- **Configurable:** Tell the component which facts you want to fetch (e.g., "Sport" and "Medals" for athletes, or "Political Party" and "Spouse" for politicians).

## Project Structure

This repository is split into two main parts:
* `/api` - The Go backend server that handles Wikipedia authentication and data scraping.
* `/client` - The frontend CSS and Javascript library
* `index.html` - A demo page showing the library in action.

## How to Test Locally

### 1. Set up the Backend (API)
1. You will need [Go](https://go.dev/doc/install) installed on your computer.
2. Open your terminal and navigate to the `api` folder
3. Rename the `example.env` to `.env` file inside the `api` folder and add your Wikimedia Enterprise credentials ([Get them here](https://enterprise.wikimedia.com/docs/#getting-started))
4. Download dependencies and start the server:
   ```bash
   go mod tidy
   go run main.go
   ```
   *You should see: `Server starting on http://localhost:8080`*

### 2. Set up the Frontend (Client)
1. Open `index.html` in your code editor.
2. Scroll to the bottom and ensure the `apiUrl` is pointing to your local Go server:
   ```javascript
   new WikiInfobox({
       apiUrl: "http://localhost:8080/api/infobox",
       selector: ".wiki-hover"
   });
   ```
3. Open `index.html` in your web browser. Hover over the names in the text to see the API fetch the data in real-time!

## 🎨 Customizing the Design

You do not need to edit `wiki-infobox.css` to change the colors or fonts. The component is powered by CSS variables. Override these variables in your own website's main stylesheet to match your style:

```css
/* Add this to your own site's CSS file */
:root {
    --wiki-bg: #ffffff;             /* Background color of the card */
    --wiki-text: #333333;           /* Standard text color */
    --wiki-title: #000000;          /* Headline and bold text color */
    --wiki-font: "Helvetica", Arial; /* Font family */
    --wiki-border-radius: 4px;      /* Sharpness of the card corners */
}
```
