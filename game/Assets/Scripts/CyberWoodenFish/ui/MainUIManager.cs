using DG.Tweening;
using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.UI;

namespace CyberWoodenFish.ui
{
    public class MainUIManager : MonoBehaviour
    {
        [SerializeField] private Button startButton;
        [SerializeField] private Button rankButton;
        [SerializeField] private AudioSource buttonAudioSource;
        [SerializeField] private float pressedScale = 0.94f;
        [SerializeField] private float pressDuration = 0.08f;
        [SerializeField] private float releaseDuration = 0.18f;

        private void Start()
        {
            startButton.onClick.AddListener(StartGame);
            rankButton.onClick.AddListener(PlayRankFeedback);
        }

        private void StartGame()
        {
            startButton.interactable = false;
            rankButton.interactable = false;
            buttonAudioSource.Play();

            Sequence sequence = PlayButtonFeedback(startButton);
            float remainingSoundDuration =
                buttonAudioSource.clip.length - pressDuration - releaseDuration;
            sequence.AppendInterval(Mathf.Max(0f, remainingSoundDuration));
            sequence.OnComplete(() => SceneManager.LoadScene("GameScene"));
        }

        private void PlayRankFeedback()
        {
            buttonAudioSource.Play();
            PlayButtonFeedback(rankButton);
        }

        private Sequence PlayButtonFeedback(Button targetButton)
        {
            Transform target = targetButton.transform;
            target.DOKill();
            target.localScale = Vector3.one;

            return DOTween.Sequence()
                .SetUpdate(true)
                .Append(target.DOScale(pressedScale, pressDuration).SetEase(Ease.OutQuad))
                .Append(target.DOScale(1f, releaseDuration).SetEase(Ease.OutBack));
        }
    }
}
