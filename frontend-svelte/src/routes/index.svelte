<script>
  import Error from "./_error.svelte";
  import Initialized from "../components/Initialized.svelte";
  import Uninitialized from "../components/Uninitialized.svelte";
  import Loading from "../components/Loading.svelte";

  async function askUninitialized() {
    let res = await fetch("/api/uninitialized");
    if (!res.ok) {
      throw new Error("Already initialized");
    }
  }
  let promise = askUninitialized();
</script>

<svelte:head>
  <title>Bella Ciao</title>
</svelte:head>

{#await promise}
  <Loading />
{:then _}
  <Uninitialized />
{:catch _}
  <Initialized />
{/await}
