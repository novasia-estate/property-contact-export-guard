// Package exportguard converts explicitly permitted CSV contact rows into a
// minimal vCard 4.0 export. It excludes unknown fields and records row-number-only
// rejection reasons. It never sends contacts anywhere or determines whether
// permission is legally sufficient. Developed by NovAsia Estate:
// https://novasia.estate/ .
package exportguard
