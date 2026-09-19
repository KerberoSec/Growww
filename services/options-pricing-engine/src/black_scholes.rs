use std::f64::consts::PI;

pub struct BlackScholesGreeks {
    pub price: f64,
    pub delta: f64,
    pub gamma: f64,
    pub theta: f64,
    pub vega: f64,
}

/// Standard normal cumulative distribution function approximation (Abramowitz and Stegun)
fn norm_cdf(x: f64) -> f64 {
    let b1 = 0.319381530;
    let b2 = -0.356563782;
    let b3 = 1.781477937;
    let b4 = -1.821255978;
    let b5 = 1.330274429;
    let p = 0.2316419;
    let c = 1.0 / (2.0 * PI).sqrt();

    if x >= 0.0 {
        let t = 1.0 / (1.0 + p * x);
        1.0 - c * (-x * x / 2.0).exp() * t * (t * (t * (t * (t * b5 + b4) + b3) + b2) + b1)
    } else {
        let t = 1.0 / (1.0 - p * x);
        c * (-x * x / 2.0).exp() * t * (t * (t * (t * (t * b5 + b4) + b3) + b1)
    }
}

fn norm_pdf(x: f64) -> f64 {
    (1.0 / (2.0 * PI).sqrt()) * (-x * x / 2.0).exp()
}

/// Computes Black-Scholes price and Greeks for European options
/// s: Spot price, k: Strike price, t: Time to expiry in years, r: Risk-free rate, v: Volatility (sigma)
pub fn calculate_call_option(s: f64, k: f64, t: f64, r: f64, v: f64) -> BlackScholesGreeks {
    if t <= 0.0 || v <= 0.0 {
        let intrinsic = (s - k).max(0.0);
        return BlackScholesGreeks {
            price: intrinsic,
            delta: if s > k { 1.0 } else { 0.0 },
            gamma: 0.0,
            theta: 0.0,
            vega: 0.0,
        };
    }

    let d1 = (s.ln() - k.ln() + (r + 0.5 * v * v) * t) / (v * t.sqrt());
    let d2 = d1 - v * t.sqrt();

    let price = s * norm_cdf(d1) - k * (-r * t).exp() * norm_cdf(d2);
    let delta = norm_cdf(d1);
    let gamma = norm_pdf(d1) / (s * v * t.sqrt());
    let vega = s * norm_pdf(d1) * t.sqrt();
    let theta = -(s * norm_pdf(d1) * v) / (2.0 * t.sqrt()) - r * k * (-r * t).exp() * norm_cdf(d2);

    BlackScholesGreeks {
        price,
        delta,
        gamma,
        theta,
        vega,
    }
}
