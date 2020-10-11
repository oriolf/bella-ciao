<script lang="ts">
  import Button from "./Button.svelte";
  import { createEventDispatcher } from "svelte";

  export let id: number;
  const dispatch = createEventDispatcher();

  async function deleteFile() {
    await fetch(`/api/users/files/delete?id=${id}`);
    dispatch("executed", true);
  }
</script>

<Button content="Delete file" type="danger" callback={() => deleteFile()} />
