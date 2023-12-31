import wasmExec_1_21_3 from "./1.21.3/wasm_exec_module.js";
import wasmExec_1_19_1 from "./1.19.1/wasm_exec_module.js";

// run takes a wasmExec object and a wasm binary URL and runs the Go program's main function.
async function run(wasmExec, wasmBinaryUrl) {
	const go = new wasmExec.Go();
	let result = await WebAssembly.instantiateStreaming(fetch(wasmBinaryUrl), go.importObject);
	go.run(result.instance);
}

async function init() {
	// calling my run() function with the wasmExec object and the wasm binary URL is like calling e.g.:
	//
	// const wasmExec_1_19_1 = new wasmExec.Go();
	// let result = await WebAssembly.instantiateStreaming(fetch("./1.19.1/main.wasm"), wasmExec_1_19_1.importObject);
	// wasmExec_1_19_1.run(result.instance);
	//
	run(wasmExec_1_19_1, "./1.19.1/main.wasm");
	run(wasmExec_1_21_3, "./1.21.3/main.wasm");
}

init();


