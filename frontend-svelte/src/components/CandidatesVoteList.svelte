<script lang="ts">
  import type { Candidate } from "../types/models.type";
  import Button from "./Buttons/Button.svelte";
  import { createEventDispatcher } from "svelte";
  import { ArrowRight } from "svelte-bootstrap-icons/lib/ArrowRight";
  import { ArrowLeft } from "svelte-bootstrap-icons/lib/ArrowLeft";
  import { ArrowDown } from "svelte-bootstrap-icons/lib/ArrowDown";
  import { ArrowUp } from "svelte-bootstrap-icons/lib/ArrowUp";

  const dispatch = createEventDispatcher();
  export let selected: boolean;
  export let list: Candidate[];
</script>

<table class="table" style="margin-bottom: 0;">
  <thead>
    <tr>
      {#if selected}
        <th />
      {/if}
      <th>Name</th>
      <th />
      {#if selected}
        <th />
      {/if}
    </tr>
  </thead>
  <tbody>
    {#each list as candidate (candidate.id)}
      <tr>
        {#if selected}
          <td>
            <Button
              content=""
              instant={() => dispatch('deselect', candidate.id)}>
              <ArrowLeft width="1.5em" height="1.5em" />
            </Button>
          </td>
        {/if}
        <td>{candidate.name}</td>
        {#if selected}
          <td>
            <Button content="" instant={() => dispatch('down', candidate.id)}>
              <ArrowDown width="1.5em" height="1.5em" />
            </Button>
          </td>
          <td>
            <Button content="" instant={() => dispatch('up', candidate.id)}>
              <ArrowUp width="1.5em" height="1.5em" />
            </Button>
          </td>
        {:else}
          <td>
            <Button content="" instant={() => dispatch('select', candidate.id)}>
              <ArrowRight width="1.5em" height="1.5em" />
            </Button>
          </td>
        {/if}
      </tr>
    {/each}
  </tbody>
</table>
