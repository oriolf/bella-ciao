<script lang="ts">
  import Error from "./_error.svelte";
  import Uninitialized from "../components/Uninitialized.svelte";
  import Loading from "../components/Loading.svelte";
  import { _ } from "svelte-i18n";

  let promise: Promise<Response> = askUninitialized();
  async function askUninitialized(): Promise<Response> {
    let res = await fetch("/api/uninitialized");
    if (!res.ok) {
      setTimeout(() => {
        window.location.href = "/";
      }, 2000);
      throw new Error("Already initialized");
    }
    return res;
  }
</script>

<svelte:head>
  <title>Bella Ciao</title>
</svelte:head>

{#await promise}
  <Loading />
{:then _}
  <Uninitialized />
{:catch e}
  <p>{$_("pages.initialize.error")}</p>
{/await}
