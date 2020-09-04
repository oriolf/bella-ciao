<script>
  import Loading from "./Loading.svelte";
  import Alert from "./Alert.svelte";

  export let rows;
  export let error;
</script>

{#await rows}
  <Loading />
{:then rws}
  <div class="card">
    <div class="card-body">
      {#each rws as row}
        <slot {row} />
      {/each}
    </div>
  </div>
{:catch _}
  <Alert type="danger" content={error} />
{/await}
