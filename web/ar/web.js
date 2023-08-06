class Argon extends HTMLElement {
    constructor() {
        super();
        const shadow = this.attachShadow({ mode: "open" });
    }
    connectedCallback() {
        const src = this.getAttribute("src");
        if (src) {
            const fullpath = new URL(src, window.location.href).href;
            importWasm(fullpath);
        }

    }
}

function importWasm(path) {
    if (importQueue.includes(path)) return;
    importQueue.push(path);
    if (wasmReady) {
        worker.postMessage({
            run: "import",
            path: path,
        });
    }
}
const importQueue = [];
customElements.define("argon-web", Argon);
const worker = new Worker("/ar/worker.js");
let wasmReady = false;
worker.onmessage = function (e) {
    switch (e.data) {
        case "ready":
            wasmReady = true;
            for (const path of importQueue) {
                worker.postMessage({
                    run: "import",
                    path: path,
                });
            }
            break;
    }
};