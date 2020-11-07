<script lang="ts">
  import { _ } from "svelte-i18n";
  import type { Election, Candidate } from "../types/models.type";
  export let election: Election;
  let candidates: Candidate[] = election.candidates.sort((x, y) => y.points - x.points);
</script>

<table class="table">
  <thead>
    <tr>
      <th>{$_('comp.election_results.name')}</th>
      <th>{$_('comp.election_results.points')}</th>
    </tr>
  </thead>
  <tbody>
    {#each candidates as candidate}
      <tr>
        <td>{candidate.name}</td>
        <td>{candidate.points}</td>
      </tr>
    {/each}
  </tbody>
</table>
