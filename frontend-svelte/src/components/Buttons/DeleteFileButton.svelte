<script lang="ts">
  import Button from "./Button.svelte";
  import { createEventDispatcher } from "svelte";
  import { _ } from "svelte-i18n";

  export let id: number;
  const dispatch = createEventDispatcher();

  async function deleteFile() {
    console.log("Deleting file ", id);
    fetch(`/api/users/files/delete?id=${id}`).then(async (res) => {
      dispatch("executed", true);
      if (!res.ok) {
        let text = await res.text();
        alert("ERROR: " + text);
      }
    });
  }
</script>

<Button content={$_("comp.buttons.delete_file")} type="danger" callback={() => deleteFile()} />
