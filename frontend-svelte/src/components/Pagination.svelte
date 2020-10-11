<script lang="ts">
  export let page: number;
  export let itemsPerPage: number;
  export let totalItems: number;
  let pages: number[];
  let lastPage: number;

  $: generatePagesArray(page);

  function generatePagesArray(page: number) {
    lastPage = Math.ceil(totalItems / itemsPerPage);
    if (lastPage <= 5) {
      pages = allPages(lastPage);
      return;
    }

    let pgs: number[] = [];
    if (page === 1) {
      pgs = [1, 2, 3, halfway(page, lastPage, true), lastPage];
    } else if (page === lastPage) {
      pgs = [1, halfway(1, page), lastPage - 2, lastPage - 1, lastPage];
    } else {
      pgs = [
        1,
        halfway(1, page, true),
        page,
        halfway(page, lastPage),
        lastPage,
      ];
    }

    pgs.sort((a, b) => a - b);
    pages = pgs;
  }

  function allPages(last: number): number[] {
    let l = [];
    for (let i = 1; i <= last; i++) {
      l.push(i);
    }
    return l;
  }

  function halfway(a: number, b: number, low = false): number {
    if (a === 1 && b === 2) {
      return 3;
    }

    if (a === lastPage - 1 && b === lastPage) {
      return lastPage - 2;
    }

    if (low) {
      return Math.ceil((a + b) / 2);
    }
    return Math.floor((a + b) / 2);
  }

  function setPage(n: number): () => void {
    return function () {
      if (n > lastPage) {
        page = lastPage;
      } else if (n <= 0) {
        page = 1;
      } else {
        page = n;
      }
    };
  }
</script>

<nav aria-label="Page navigation example">
  <ul class="pagination">
    <li class="page-item" class:disabled={page === 1 || totalItems === 0}>
      <a class="page-link" href="." on:click={setPage(page - 1)}>Previous</a>
    </li>

    {#each pages as pg}
      {#if pg}
        <li class="page-item" class:active={pg === page}>
          <a class="page-link" href="." on:click={setPage(pg)}>{pg}</a>
        </li>
        {#if pg === page}<span class="sr-only">(current)</span>{/if}
      {:else}
        <li class="page-item disabled">
          <a class="page-link" href=".">...</a>
        </li>
      {/if}
    {/each}

    <li
      class="page-item"
      class:disabled={page === lastPage || totalItems === 0}>
      <a class="page-link" href="." on:click={setPage(page + 1)}>Next</a>
    </li>
  </ul>
</nav>
