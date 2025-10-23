This is an example of how I used tanstack router's query params api to render different "screens"
```tsx
import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import {
  ArrowUpDown,
  Eye,
  ListOrdered,
  PlusIcon,
  Sparkles,
  Trash2,
} from 'lucide-react';
import Masonry from 'react-masonry-css';

import { DeleteProjectDialog } from './-components/delete-project-dialog';
import { EditProjectForm } from './-components/edit-project-form';
import { ProjectImagePreview } from './-components/project-image-preview';
import { ProjectOptionsMenu } from './-components/project-options-menu';
import { ProjectOrderForm } from './-components/project-order-form';
import { ProjectUploadForm } from './-components/project-upload-form';

import { categoriesSelectQueryOptions } from '@/lib/categories.queries';
import { projectQueryOptions } from '@/lib/projects.queries';

const projectOptionsSchema = z.object({
  upload: z.boolean().default(false),
  order: z.boolean().default(false),
  edit: z.boolean().default(false),
  delete: z.boolean().default(false),
  preview: z.boolean().default(false),
});

export const Route = createFileRoute('/_auth/projects/$projectId')({
  params: {
    parse: (params) => ({
      projectId: z.coerce.number().int().parse(params.projectId),
    }),
  },
  loader(opts) {
    opts.context.queryClient.ensureQueryData(
      projectQueryOptions(opts.params.projectId)
    );
    opts.context.queryClient.ensureQueryData(categoriesSelectQueryOptions());
  },
  component: RouteComponent,
  validateSearch: projectOptionsSchema,
});

function RouteComponent() {
  const params = Route.useParams();
  const {
    upload,
    order,
    edit,
    delete: deleteAction,
    preview,
  } = Route.useSearch();
  const { data: project } = useSuspenseQuery(
    projectQueryOptions(params.projectId)
  );
  const { data: categories } = useSuspenseQuery(categoriesSelectQueryOptions());

  if (upload) {
    return <ProjectUploadForm id={project.id} />;
  }

  if (order) {
    return <ProjectOrderForm projectImages={project.images} />;
  }

  if (edit) {
    return <EditProjectForm project={project} categories={categories} />;
  }

  if (deleteAction) {
    return <DeleteProjectDialog isOpen={deleteAction} id={project.id} />;
  }

  if (preview) {
    return <ProjectImagePreview project={project} />;
  }

  return (
    <section className="container space-y-12 p-6">
      <div className="flex flex-col gap-4">
        <h1 className="scroll-m-20 pb-2 text-3xl font-bold leading-relaxed first:mt-0">
          {project.name}
        </h1>
        <nav className="flex gap-4">
          <Button asChild>
            <Link to="." search={{ upload: true }}>
              <PlusIcon className="mr-2 h-4 w-4" /> Upload
            </Link>
          </Button>
          <Button asChild variant="outline">
            <Link to="." search={{ order: true }}>
              <ArrowUpDown className="mr-2 h-4 w-4" /> Order
            </Link>
          </Button>
          <Button asChild variant="outline">
            <Link to="." search={{ preview: true }}>
              <Eye className="mr-2 h-4 w-4" /> Preview
            </Link>
          </Button>
          <Button asChild variant="outline">
            <Link to="." search={{ edit: true }}>
              <ListOrdered className="mr-2 h-4 w-4" /> Edit
            </Link>
          </Button>
          <Button asChild variant="destructive">
            <Link to="." search={{ delete: true }}>
              <Trash2 className="mr-2 h-4 w-4" /> Delete
            </Link>
          </Button>
        </nav>
      </div>
      {project.images.length ? (
        <Masonry
          breakpointCols={{
            default: 4,
            1280: 4,
            1024: 3,
            768: 2,
            640: 1,
          }}
          className="ml-[-2rem] flex w-auto"
          columnClassName="pl-[2rem] bg-clip-border"
        >
          {project.images.map((image) => {
            return (
              <div key={image.id} className="mb-[2rem]">
                <div className="group relative overflow-hidden rounded-lg">
                  <img
                    src={image.thumbnailUrl}
                    alt={image.description ?? image.name ?? 'image'}
                  />
                  <ProjectOptionsMenu
                    image={{
                      id: image.id,
                      name: image.name!,
                      url: image.url,
                    }}
                    project={{
                      id: project.id,
                      thumbnailId: project.thumbnailId,
                    }}
                  />
                  {project.thumbnailId === image.id && (
                    <div className="absolute bottom-2 right-2">
                      <Sparkles
                        className="text-yellow-400"
                        size={32}
                        fill="rgb(253 224 71)"
                      />
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </Masonry>
      ) : (
        <p>No images</p>
      )}
    </section>
  );
}
```