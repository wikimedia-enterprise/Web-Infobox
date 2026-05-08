/* client/wiki-infobox.js */
class WikiInfobox {
    constructor(options = {}) {
        this.apiUrl = options.apiUrl || 'http://localhost:8080/api/infobox';
        this.selector = options.selector || '.wiki-hover';
        this.desiredFacts = options.desiredFacts || [
            "Full name", "Nickname", "Born", "Nationality", "Sport", "Medal record", "Personal best(s)"
        ];

        this.injectHTML();
        this.bindEvents();
    }

    injectHTML() {
        if (document.getElementById('wiki-tooltip')) return;
        const html = `
            <div id="wiki-tooltip">
                <div id="wiki-loading">Loading Wikipedia Data...</div>
                <div id="wiki-content-wrapper" style="display: none;">
                    <div class="wiki-left">
                        <div id="wiki-image-container">
                            <img id="wiki-image" src="" alt="Wikipedia Image">
                            <div id="wiki-image-author"></div>
                        </div>
                        <div id="wiki-caption"></div>
                    </div>
                    <div class="wiki-right">
                        <h3 id="wiki-title"></h3>
                        <div class="wiki-fact-grid" id="wiki-facts"></div>
                        <div class="wiki-footer">
                            <div style="display: flex; align-items: center; justify-content: center; gap: 6px; margin-bottom: 6px;">
                                <span class="wiki-footer-text" style="line-height: 1;">From</span>
                                <img src="https://upload.wikimedia.org/wikipedia/commons/b/bb/Wikipedia_wordmark.svg" alt="Wikipedia" class="wiki-wordmark">
                                <span class="wiki-footer-text" style="margin-left: 4px;">Content under CC BY-SA 4.0</span>
                            </div>
                            <div style="display: flex; align-items: center; justify-content: center; gap: 12px;">
                                <span id="wiki-last-updated" class="wiki-footer-text"></span>
                                <a id="wiki-link" class="wiki-footer-link" href="#" target="_blank">Read source article</a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>`;
        document.body.insertAdjacentHTML('beforeend', html);

        this.tooltip = document.getElementById('wiki-tooltip');
        this.titleEl = document.getElementById('wiki-title');
        this.captionEl = document.getElementById('wiki-caption');
        this.imgContainer = document.getElementById('wiki-image-container');
        this.imgEl = document.getElementById('wiki-image');
        this.authorEl = document.getElementById('wiki-image-author');
        this.loadingEl = document.getElementById('wiki-loading');
        this.contentWrapper = document.getElementById('wiki-content-wrapper');
        this.factsContainer = document.getElementById('wiki-facts');
        this.linkEl = document.getElementById('wiki-link');
        this.lastUpdatedEl = document.getElementById('wiki-last-updated');
    }

    bindEvents() {
        let hoverTimeout, hideTimeout;
        document.querySelectorAll(this.selector).forEach(link => {
            link.addEventListener('mouseenter', (e) => {
                clearTimeout(hideTimeout);
                const topic = e.target.getAttribute('data-topic');
                hoverTimeout = setTimeout(() => this.showTooltip(e.target, topic), 400);
            });
            link.addEventListener('mouseleave', () => {
                clearTimeout(hoverTimeout);
                hideTimeout = setTimeout(() => this.tooltip.classList.remove('visible'), 300);
            });
        });

        this.tooltip.addEventListener('mouseenter', () => clearTimeout(hideTimeout));
        this.tooltip.addEventListener('mouseleave', () => hideTimeout = setTimeout(() => this.tooltip.classList.remove('visible'), 300));
    }

    async showTooltip(element, topic) {
        const rect = element.getBoundingClientRect();
        this.tooltip.style.left = `${rect.left + window.scrollX}px`;
        this.tooltip.style.top = `${rect.bottom + window.scrollY + 8}px`;

        this.loadingEl.style.display = 'block';
        this.contentWrapper.style.display = 'none';
        this.tooltip.classList.add('visible');

        try {
            const response = await fetch(`${this.apiUrl}?topic=${encodeURIComponent(topic)}`);
            if (!response.ok) throw new Error('Not found');
            const data = await response.json();

            this.titleEl.textContent = data.title;
            this.captionEl.textContent = data.caption || "";
            data.facts["Full name"] = data.facts["Full name"] || data.title;

            if (data.image_url) {
                this.imgEl.src = data.image_url;
                this.imgContainer.style.display = 'block';
                if (data.image_author) {
                    this.authorEl.textContent = `Image by ${data.image_author} - CC BY-SA 4.0`;
                    this.authorEl.style.display = 'block';
                } else {
                    this.authorEl.style.display = 'none';
                }
            } else {
                this.imgContainer.style.display = 'none';
            }

            if (data.url) {
                this.linkEl.href = data.url;
                this.linkEl.style.display = 'inline-block';
            } else {
                this.linkEl.style.display = 'none';
            }

            this.lastUpdatedEl.textContent = data.last_updated ? `Last updated ${data.last_updated}` : '';

            this.factsContainer.innerHTML = '';
            this.desiredFacts.forEach(key => {
                let val = data.facts[key];
                if (val) {
                    if (key === "Born") val = val.replace(/\s*\(.*?\)/g, '');
                    val = val.replace(/\bSt\b /g, 'St. ');
                    let isStacked = (key === "Medal record" || key.includes("Personal best")) ? " stacked" : "";
                    this.factsContainer.innerHTML += `<div class="wiki-fact${isStacked}"><strong>${key}</strong><span>${val}</span></div>`;
                }
            });

            this.loadingEl.style.display = 'none';
            this.contentWrapper.style.display = 'flex';
        } catch (error) {
            this.loadingEl.textContent = "Could not load Wikipedia data.";
        }
    }
}
