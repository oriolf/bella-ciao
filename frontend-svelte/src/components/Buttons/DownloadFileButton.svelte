<script lang="ts">
  import Button from "./Button.svelte";
  import { _ } from "svelte-i18n";

  export let id: number;
  export let filename: string;

  async function downloadFile() {
    let res = await fetch(`/api/users/files/download?id=${id}`);
    let content = window.URL.createObjectURL(await res.blob());
    var a = document.createElement("a");
    a.href = content;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
  }
</script>

<Button content={$_("comp.buttons.download_file")} callback={() => downloadFile()} />
