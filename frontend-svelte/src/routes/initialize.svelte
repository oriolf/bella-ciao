<script>
  import Error from "./_error.svelte";
  import Uninitialized from "../components/Uninitialized.svelte";
  import Loading from "../components/Loading.svelte";

  let promise = askUninitialized();
  async function askUninitialized() {
    let res = await fetch("/api/uninitialized");
    if (!res.ok) {
      setTimeout(() => {
        window.location = "/";
      }, 2000);
      throw new Error("Already initialized");
    }
  }
</script>

<svelte:head>
  <title>Bella Ciao</title>
</svelte:head>

{#await promise}
  <Loading />
{:then _}
  <Uninitialized />
{:catch _}
  <p>Already initialized!</p>
{/await}
