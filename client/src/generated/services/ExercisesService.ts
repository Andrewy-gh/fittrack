/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { exercise_CreateExerciseRequest } from '../models/exercise_CreateExerciseRequest';
import type { exercise_CreateExerciseResponse } from '../models/exercise_CreateExerciseResponse';
import type { exercise_ExerciseResponse } from '../models/exercise_ExerciseResponse';
import type { exercise_ExerciseWithSetsResponse } from '../models/exercise_ExerciseWithSetsResponse';
import type { exercise_RecentSetsResponse } from '../models/exercise_RecentSetsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ExercisesService {
    /**
     * List exercises
     * Get all exercises for the authenticated user
     * @returns exercise_ExerciseResponse OK
     * @throws ApiError
     */
    public static getExercises(): CancelablePromise<Array<exercise_ExerciseResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/exercises',
            errors: {
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Get or create exercise
     * Get an existing exercise by name or create it if it doesn't exist
     * @param request Exercise data
     * @returns exercise_CreateExerciseResponse OK
     * @throws ApiError
     */
    public static postExercises(
        request: exercise_CreateExerciseRequest,
    ): CancelablePromise<exercise_CreateExerciseResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/exercises',
            body: request,
            errors: {
                400: `Bad Request`,
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Get exercise with sets
     * Get a specific exercise with all its sets from workouts. Returns empty array when exercise has no sets.
     * @param id Exercise ID
     * @returns exercise_ExerciseWithSetsResponse Success (may be empty array)
     * @throws ApiError
     */
    public static getExercises1(
        id: number,
    ): CancelablePromise<Array<exercise_ExerciseWithSetsResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/exercises/{id}',
            path: {
                'id': id,
            },
            errors: {
                400: `Bad Request`,
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Get recent sets for exercise
     * Get the 3 most recent sets for a specific exercise. Returns empty array when exercise has no sets.
     * @param id Exercise ID
     * @returns exercise_RecentSetsResponse Success (may be empty array)
     * @throws ApiError
     */
    public static getExercisesRecentSets(
        id: number,
    ): CancelablePromise<Array<exercise_RecentSetsResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/exercises/{id}/recent-sets',
            path: {
                'id': id,
            },
            errors: {
                400: `Bad Request`,
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
}
