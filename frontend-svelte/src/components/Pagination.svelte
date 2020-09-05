<script>
  import Alert from "./Alert.svelte";

  export let page;
  export let itemsPerPage;
  export let totalItems;
  let pages;
  let lastPage;
  $: generatePagesArray(page);

  function generatePagesArray(page) {
    lastPage = Math.ceil(totalItems / itemsPerPage);
    if (lastPage <= 5) {
      pages = allPages(lastPage);
      return;
    }

    let pgs = [];
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

  function allPages(last) {
    let l = [];
    for (let i = 1; i <= last; i++) {
      l.push(i);
    }
    return l;
  }

  function halfway(a, b, low) {
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

  function setPage(n) {
    return function () {
      page = n;
    };
  }
</script>

<nav aria-label="Page navigation example">
  <ul class="pagination">
    <li class="page-item" class:disabled={page === 1}>
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

    <li class="page-item" class:disabled={page === lastPage}>
      <a class="page-link" href="." on:click={setPage(page + 1)}>Next</a>
    </li>
  </ul>
</nav>
