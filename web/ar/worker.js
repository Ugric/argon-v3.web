importScripts("/ar/wasm_exec.js");

onmessage = (e) => {
    switch (e.data.run) {
        case "import":
            Ar.import(e.data.path);
            break;
    }
};

const go = new Go();
WebAssembly.instantiateStreaming(
    fetch("/ar/bin/argon.wasm"),
    go.importObject
).then((result) => {
    postMessage("ready");
    go.run(result.instance);
});
