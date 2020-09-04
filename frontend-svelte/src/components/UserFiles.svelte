<script>
  import CardTable from "./CardTable.svelte";
  import DownloadFileButton from "./Buttons/DownloadFileButton.svelte";
  import DeleteFileButton from "./Buttons/DeleteFileButton.svelte";
  import { get, sortByField } from "../util.js";

  let files;
  export let reloadFiles = 0;
  $: getFiles(reloadFiles);

  async function getFiles(_) {
    files = get("/api/users/files/own", sortByField("name"));
  }
</script>

<CardTable
  headers={['Description', '', '']}
  rows={files}
  error="Could not obtain the uploaded files"
  let:row>
  <td>{row.description}</td>
  <td>
    <DownloadFileButton id={row.id} filename={row.name} />
  </td>
  <td>
    <DeleteFileButton on:executed={getFiles} />
  </td>
</CardTable>
