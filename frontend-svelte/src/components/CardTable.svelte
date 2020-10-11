<script lang="ts">
  import Loading from "./Loading.svelte";
  import Alert from "./Alert.svelte";

  export let headers: string[];
  export let rows: Promise<any[]>;
  export let error: string;
</script>

{#await rows}
  <Loading />
{:then rws}
  {#if rws.length > 0}
    <div class="card" style="padding: 5px; margin-bottom: 10px;">
      <table class="table" style="margin-bottom: 0;">
        <thead>
          <tr>
            {#each headers as header}
              <th>{header}</th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each rws as row}
            <tr>
              <slot {row} />
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
{:catch _}
  <Alert type="danger" content={error} />
{/await}
