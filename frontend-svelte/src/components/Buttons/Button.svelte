<script>
  import Loading from "../Loading.svelte";

  export let content;
  export let callback;
  export let type = "primary";
  let promise;
  let classes = `align-middle btn btn-sm btn-outline-${type}`;

  function handleClick() {
    promise = callback();
  }
</script>

<button
  type="button"
  class={classes}
  style="width: 100%;"
  on:click={handleClick}>
  {#await promise}
    <Loading />
  {:then _}
    {content}
  {:catch _}
    {content}
  {/await}
</button>
