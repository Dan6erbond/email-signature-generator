function copySignature() {
  const signature = document.getElementById("signature");

  try {
    navigator.clipboard.write([
      new ClipboardItem({
        "text/html": new Blob([signature.innerHTML], { type: "text/html" }),
      }),
    ]);
    console.log("Copied to clipboard successfully!");
  } catch (err) {
    console.error("Failed to copy: ", err);
  }
}

function copySignatureHTML() {
  const signature = document.getElementById("signature");

  try {
    navigator.clipboard.writeText(signature.innerHTML);
    console.log("Copied to clipboard successfully!");
  } catch (err) {
    console.error("Failed to copy: ", err);
  }
}
