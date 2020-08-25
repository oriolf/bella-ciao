<script>
  import Button from "./Button.svelte";
  export let url;
  export let filename;

  async function downloadFile() {
    let res = await fetch(url);
    let content = window.URL.createObjectURL(await res.blob());
    var a = document.createElement("a");
    a.href = content;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
  }
</script>

<Button content="Download file" callback={() => downloadFile()} />
