<script lang="ts">
  import Loading from "../Loading.svelte";

  export let content: string;
  export let callback: () => Promise<any>;
  export let instant: () => void;
  export let type: string = "primary";
  export let disabled = false;
  let promise: Promise<any>;
  let classes: string = `align-middle btn btn-sm btn-outline-${type}`;

  function handleClick() {
    if (instant) {
      instant();
      return;
    }
    promise = callback();
  }
</script>

<button
  type="button"
  class={classes}
  style="width: 100%;"
  on:click={handleClick}
  disabled={disabled}>
  {#await promise}
    <Loading />
  {:then _}
    {content}
  {:catch _}
    {content}
  {/await}
  <slot />
</button>
