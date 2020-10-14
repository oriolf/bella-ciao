<script lang="ts">
  import Button from "./Button.svelte";
  import { createEventDispatcher } from "svelte";
  import { _ } from "svelte-i18n";

  export let id: number;
  const dispatch = createEventDispatcher();

  async function deleteFile() {
    await fetch(`/api/users/files/delete?id=${id}`);
    dispatch("executed", true);
  }
</script>

<Button content={$_("comp.buttons.delete_file")} type="danger" callback={() => deleteFile()} />
