interface IDistProps {
    Min: number;
    Max: number;
    Mean: number;
    Variance: number;
}

type DistRez = ( () => number /*| IDistProps */ );

declare namespace Prob {
    function uniform(min: number, max: number): DistRez;
    function normal(mu: number, sigma: number): DistRez;
    function exponential(lambda: number): DistRez;
    function lognormal(mu: number, sigma: number): DistRez;
    function poisson(lambda: number): DistRez;
    function zipf(s: number, N: number): DistRez;
}