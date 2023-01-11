VERSION=$1
GH_PAGES_BRANCH=gh_pages_$VERSION

cd ~
helm package aqueduct-helm/

cd ~/aqueduct-helm/
git checkout gh-pages && git pull origin gh-pages
git branch $GH_PAGES_BRANCH && git checkout $GH_PAGES_BRANCH
mv ~/aqueduct-$VERSION.tgz .

cd ~ && helm repo index aqueduct-helm/ --url https://aqueducthq.github.io/aqueduct-helm/
cd ~/aqueduct-helm
git add --all && git commit -m "prepare gh page for release"
echo "Please make a PR to merge the following branch to gh-pages:"
echo $GH_PAGES_BRANCH