<script>
  import CardTable from "./CardTable.svelte";
  import DownloadFileButton from "./DownloadFileButton.svelte";
  import Button from "./Button.svelte";
  import { get, sortByField } from "../util.js";

  let files;
  export let reloadFiles = 0;
  $: getFiles(reloadFiles);

  async function getFiles(_) {
    files = get("/api/users/files/own", sortByField("name"));
  }

  async function deleteFile(id) {
    await fetch(`/api/users/files/delete?id=${id}`);
    getFiles();
  }
</script>

<CardTable
  headers={['Description', '', '']}
  rows={files}
  error="Could not obtain the uploaded files"
  let:row>
  <td>{row.description}</td>
  <td>
    <DownloadFileButton
      url={`/api/users/files/download?id=${row.id}`}
      filename={row.name} />
  </td>
  <td>
    <Button
      content="Delete file"
      type="danger"
      callback={() => deleteFile(row.id)} />
  </td>
</CardTable>
