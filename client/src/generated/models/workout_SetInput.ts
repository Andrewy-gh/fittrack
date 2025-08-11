/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type workout_SetInput = {
    reps: number;
    setType: workout_SetInput.setType;
    weight?: number;
};
export namespace workout_SetInput {
    export enum setType {
        WARMUP = 'warmup',
        WORKING = 'working',
    }
}

