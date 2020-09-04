<script>
  export let page;
  export let itemsPerPage;
  export let totalItems;
  let pages;
  let lastPage;
  $: generatePagesArray(page);

  function generatePagesArray(page) {
    pages = [1, 3, 5, 7, 9];
    lastPage = Math.ceil(totalItems / itemsPerPage);
    let pgs = [];
    let aux = [1, halfway(1, page), page, halfway(page, lastPage), lastPage];
    aux.sort((a, b) => a - b);
    for (let x of aux) {
      if (pgs.length === 0 || x !== pgs[pgs.length - 1]) {
        pgs.push(x);
      }
    }
    pages = pgs;
  }

  function halfway(a, b) {
    if (a >= b) {
      return a;
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
