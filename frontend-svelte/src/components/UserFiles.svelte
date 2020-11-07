<script lang="ts">
  import CardTable from "./CardTable.svelte";
  import DownloadFileButton from "./Buttons/DownloadFileButton.svelte";
  import DeleteFileButton from "./Buttons/DeleteFileButton.svelte";
  import { get, sortByField } from "../util";
  import type { UserFile } from "../types/models.type";
import { _ } from "svelte-i18n";

  let files: Promise<UserFile[]>;
  export let reloadFiles: number = 0;
  $: getFiles(reloadFiles);

  async function getFiles(_) {
    files = get("/api/users/files/own", sortByField("name"));
  }
</script>

<CardTable
  headers={[$_("comp.user_files.description"), '', '']}
  rows={files}
  error={$_("comp.user_files.files_err")}
  let:row>
  <td>{row.description}</td>
  <td>
    <DownloadFileButton id={row.id} filename={row.name} />
  </td>
  <td>
    <DeleteFileButton id={row.id} on:executed={getFiles} />
  </td>
</CardTable>
