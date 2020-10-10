<script lang="ts">
  import type { Candidate } from "../types/models.type";
  import Button from "./Buttons/Button.svelte";
  import { createEventDispatcher } from "svelte";

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
              content="deselect"
              instant={() => dispatch('deselect', candidate.id)} />
          </td>
        {/if}
        <td>{candidate.name}</td>
        {#if selected}
          <td>
            <Button
              content="down"
              instant={() => dispatch('down', candidate.id)} />
          </td>
          <td>
            <Button content="up" instant={() => dispatch('up', candidate.id)} />
          </td>
        {:else}
          <td>
            <Button
              content="select"
              instant={() => dispatch('select', candidate.id)} />
          </td>
        {/if}
      </tr>
    {/each}
  </tbody>
</table>
