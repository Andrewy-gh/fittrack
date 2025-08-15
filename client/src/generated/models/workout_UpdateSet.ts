/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type workout_UpdateSet = {
    reps: number;
    setType: workout_UpdateSet.setType;
    weight?: number;
};
export namespace workout_UpdateSet {
    export enum setType {
        WARMUP = 'warmup',
        WORKING = 'working',
    }
}

