/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { response_SuccessResponse } from '../models/response_SuccessResponse';
import type { workout_CreateWorkoutRequest } from '../models/workout_CreateWorkoutRequest';
import type { workout_UpdateWorkoutRequest } from '../models/workout_UpdateWorkoutRequest';
import type { workout_WorkoutResponse } from '../models/workout_WorkoutResponse';
import type { workout_WorkoutWithSetsResponse } from '../models/workout_WorkoutWithSetsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class WorkoutsService {
    /**
     * List workouts
     * Get all workouts for the authenticated user
     * @returns workout_WorkoutResponse OK
     * @throws ApiError
     */
    public static getWorkouts(): CancelablePromise<Array<workout_WorkoutResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/workouts',
            errors: {
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Create a new workout
     * Create a new workout with exercises and sets
     * @param request Workout data
     * @returns response_SuccessResponse OK
     * @throws ApiError
     */
    public static postWorkouts(
        request: workout_CreateWorkoutRequest,
    ): CancelablePromise<response_SuccessResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/workouts',
            body: request,
            errors: {
                400: `Bad Request`,
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Get workout with sets
     * Get a specific workout with all its sets and exercises
     * @param id Workout ID
     * @returns workout_WorkoutWithSetsResponse OK
     * @throws ApiError
     */
    public static getWorkouts1(
        id: number,
    ): CancelablePromise<Array<workout_WorkoutWithSetsResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/workouts/{id}',
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
     * Update an existing workout
     * Update workout metadata (date/notes) for the authenticated user. Returns 204 No Content on success.
     * @param id Workout ID
     * @param request Updated workout data
     * @returns void
     * @throws ApiError
     */
    public static putWorkouts(
        id: number,
        request: workout_UpdateWorkoutRequest,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/workouts/{id}',
            path: {
                'id': id,
            },
            body: request,
            errors: {
                400: `Bad Request - Invalid input or validation error`,
                401: `Unauthorized - Invalid token`,
                404: `Not Found - Workout not found or doesn't belong to user`,
                500: `Internal Server Error`,
            },
        });
    }
}
