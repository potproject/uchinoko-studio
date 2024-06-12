const random = true;

export const getID = () => {
    if (!random) {
        return "guest"
    }
    // localStorageからIDを取得
    const id = localStorage.getItem('id');
    // IDが存在しない場合ランダムなIDを生成
    if (id === null) {
        const newId = Math.random().toString(32).substring(2);
        localStorage.setItem('id', newId);
        return newId;
    }
    return id;
}

