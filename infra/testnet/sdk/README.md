# SDK Publishing Checklist

This repo already contains SDKs for Go, JavaScript, and Python. To meet
community expectations, publish the packages to their registries after tagging
the release.

## JavaScript (npm)

```bash
cd sdk/javascript
npm publish --access public
```

## Python (PyPI)

```bash
cd sdk/python
python -m build
twine upload dist/*
```

## Go (pkg.go.dev)

Go modules publish automatically when you tag releases in git.

```bash
git tag sdk-go-vX.Y.Z
git push origin sdk-go-vX.Y.Z
```
