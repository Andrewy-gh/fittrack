export type PaginationItem = number | "ellipsis";

export function getPaginationItems(
  currentPage: number,
  totalPages: number,
): PaginationItem[] {
  const safeTotalPages = Math.max(1, totalPages);
  const safeCurrentPage = Math.min(
    Math.max(1, currentPage),
    safeTotalPages,
  );

  if (safeTotalPages <= 5) {
    return Array.from({ length: safeTotalPages }, (_, index) => index + 1);
  }

  let pages: number[];

  if (safeCurrentPage <= 3) {
    pages = [1, 2, 3, 4, safeTotalPages];
  } else if (safeCurrentPage >= safeTotalPages - 2) {
    pages = [
      1,
      safeTotalPages - 3,
      safeTotalPages - 2,
      safeTotalPages - 1,
      safeTotalPages,
    ];
  } else {
    pages = [
      1,
      safeCurrentPage - 1,
      safeCurrentPage,
      safeCurrentPage + 1,
      safeTotalPages,
    ];
  }

  const items: PaginationItem[] = [];

  pages.forEach((page, index) => {
    if (index === 0) {
      items.push(page);
      return;
    }

    const previousPage = pages[index - 1];
    if (page - previousPage > 1) {
      items.push("ellipsis");
    }

    items.push(page);
  });

  return items;
}
